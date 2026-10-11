package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/shared/apperror"
)

const maxDuplicateAttempts = 100

type Repository interface {
	Create(ctx context.Context, m domain.Endpoint) (domain.Endpoint, error)
	List(ctx context.Context) ([]domain.Endpoint, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.Endpoint, error)
	GetByID(ctx context.Context, id string) (domain.Endpoint, error)
	FindMatch(ctx context.Context, method, path string) (domain.Endpoint, error)
	FindMatchByProject(ctx context.Context, projectID, method, path string) (domain.Endpoint, error)
	Update(ctx context.Context, m domain.Endpoint) (domain.Endpoint, error)
	Delete(ctx context.Context, id string) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.Endpoint, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Endpoint{}, apperror.New(apperror.Invalid, "invalid id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]domain.Endpoint, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]domain.Endpoint, error) {
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) Match(ctx context.Context, method, path string) (domain.Endpoint, error) {
	cleanMethod := strings.ToUpper(strings.TrimSpace(method))
	cleanPath := domain.NormalizePath(path)
	m, err := s.repo.FindMatch(ctx, cleanMethod, cleanPath)
	if err != nil {
		if apperror.Is(err, apperror.NotFound) {
			return domain.Endpoint{}, apperror.New(apperror.NotFound, "no mock endpoint configured")
		}
		return domain.Endpoint{}, err
	}
	return m, nil
}

func (s *Service) MatchProject(ctx context.Context, projectID, method, path string) (domain.Endpoint, error) {
	cleanMethod := strings.ToUpper(strings.TrimSpace(method))
	cleanPath := domain.NormalizePath(path)
	m, err := s.repo.FindMatchByProject(ctx, projectID, cleanMethod, cleanPath)
	if err != nil {
		if apperror.Is(err, apperror.NotFound) {
			return domain.Endpoint{}, apperror.New(apperror.NotFound, "no mock endpoint configured")
		}
		return domain.Endpoint{}, err
	}
	return m, nil
}

func (s *Service) CreateWithProject(ctx context.Context, projectID *string, method, path string, status int, response json.RawMessage) (domain.Endpoint, error) {
	cleanMethod, err := domain.ValidateMethod(method)
	if err != nil {
		return domain.Endpoint{}, err
	}
	cleanPath, err := domain.ValidatePath(path)
	if err != nil {
		return domain.Endpoint{}, err
	}
	cleanStatus, err := domain.ValidateStatus(status)
	if err != nil {
		return domain.Endpoint{}, err
	}
	cleanResponse, err := domain.ValidateResponse(response)
	if err != nil {
		return domain.Endpoint{}, err
	}

	var existing domain.Endpoint
	if projectID != nil && *projectID != "" {
		existing, err = s.repo.FindMatchByProject(ctx, *projectID, cleanMethod, cleanPath)
	} else {
		existing, err = s.repo.FindMatch(ctx, cleanMethod, cleanPath)
	}
	if err != nil && !apperror.Is(err, apperror.NotFound) {
		return domain.Endpoint{}, err
	}
	if err == nil && existing.ID != "" {
		return domain.Endpoint{}, apperror.New(apperror.Conflict,
			fmt.Sprintf("endpoint for %s %s already exists", cleanMethod, cleanPath))
	}

	var pid *string
	if projectID != nil && *projectID != "" {
		cleanPid := strings.Clone(*projectID)
		pid = &cleanPid
	}

	m := domain.Endpoint{
		ID:        uuid.NewString(),
		ProjectID: pid,
		Method:    cleanMethod,
		Path:      cleanPath,
		Status:    cleanStatus,
		Response:  cleanResponse,
		CreatedAt: time.Now().UTC(),
	}

	return s.repo.Create(ctx, m)
}

func (s *Service) Create(ctx context.Context, method, path string, status int, response json.RawMessage) (domain.Endpoint, error) {
	return s.CreateWithProject(ctx, nil, method, path, status, response)
}

func (s *Service) Update(ctx context.Context, id, method, path string, status int, response json.RawMessage) (domain.Endpoint, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Endpoint{}, apperror.New(apperror.Invalid, "invalid id")
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Endpoint{}, err
	}

	cleanMethod := existing.Method
	if strings.TrimSpace(method) != "" {
		m, err := domain.ValidateMethod(method)
		if err != nil {
			return domain.Endpoint{}, err
		}
		cleanMethod = m
	}

	cleanPath := existing.Path
	if strings.TrimSpace(path) != "" {
		p, err := domain.ValidatePath(path)
		if err != nil {
			return domain.Endpoint{}, err
		}
		cleanPath = p
	}

	cleanStatus := existing.Status
	if status != 0 {
		st, err := domain.ValidateStatus(status)
		if err != nil {
			return domain.Endpoint{}, err
		}
		cleanStatus = st
	}

	cleanResponse := existing.Response
	if len(response) > 0 {
		resp, err := domain.ValidateResponse(response)
		if err != nil {
			return domain.Endpoint{}, err
		}
		cleanResponse = resp
	}

	var other domain.Endpoint
	if existing.ProjectID != nil {
		other, err = s.repo.FindMatchByProject(ctx, *existing.ProjectID, cleanMethod, cleanPath)
	} else {
		other, err = s.repo.FindMatch(ctx, cleanMethod, cleanPath)
	}
	if err != nil && !apperror.Is(err, apperror.NotFound) {
		return domain.Endpoint{}, err
	}
	if err == nil && other.ID != "" && other.ID != id {
		return domain.Endpoint{}, apperror.New(apperror.Conflict,
			fmt.Sprintf("endpoint for %s %s already exists", cleanMethod, cleanPath))
	}

	existing.Method = cleanMethod
	existing.Path = cleanPath
	existing.Status = cleanStatus
	existing.Response = cleanResponse

	return s.repo.Update(ctx, existing)
}

func (s *Service) Duplicate(ctx context.Context, id string) (domain.Endpoint, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Endpoint{}, apperror.New(apperror.Invalid, "invalid id")
	}
	src, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Endpoint{}, err
	}
	base := src.Path + "-copy"
	for i := 1; i <= maxDuplicateAttempts; i++ {
		if err := ctx.Err(); err != nil {
			return domain.Endpoint{}, err
		}
		path := base
		if i > 1 {
			path = fmt.Sprintf("%s-%d", base, i)
		}
		dup := domain.Endpoint{
			ID:        uuid.NewString(),
			ProjectID: src.ProjectID,
			Method:    src.Method,
			Path:      path,
			Status:    src.Status,
			Response:  src.Response,
			CreatedAt: time.Now().UTC(),
		}
		created, err := s.repo.Create(ctx, dup)
		if apperror.Is(err, apperror.Conflict) {
			continue
		}
		if err != nil {
			return domain.Endpoint{}, err
		}
		return created, nil
	}
	return domain.Endpoint{}, apperror.New(apperror.Conflict, "could not generate unique duplicate path")
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return apperror.New(apperror.Invalid, "invalid id")
	}
	return s.repo.Delete(ctx, id)
}

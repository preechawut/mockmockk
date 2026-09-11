package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/platform/apperror"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
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
	if err != nil && !errors.Is(err, domain.ErrNoMatch) {
		return domain.Endpoint{}, err
	}
	if err == nil && existing.ID != "" {
		return domain.Endpoint{}, apperror.Conflict("CONFLICT",
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

func (s *Service) List(ctx context.Context) ([]domain.Endpoint, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]domain.Endpoint, error) {
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.Endpoint, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Endpoint{}, domain.ErrInvalidInput
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id, method, path string, status int, response json.RawMessage) (domain.Endpoint, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Endpoint{}, domain.ErrInvalidInput
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

	other, err := s.repo.FindMatch(ctx, cleanMethod, cleanPath)
	if err != nil && !errors.Is(err, domain.ErrNoMatch) {
		return domain.Endpoint{}, err
	}
	if err == nil && other.ID != "" && other.ID != id {
		return domain.Endpoint{}, apperror.Conflict("CONFLICT",
			fmt.Sprintf("endpoint for %s %s already exists", cleanMethod, cleanPath))
	}

	existing.Method = cleanMethod
	existing.Path = cleanPath
	existing.Status = cleanStatus
	existing.Response = cleanResponse

	return s.repo.Update(ctx, existing)
}

const maxDuplicateAttempts = 100

func (s *Service) Duplicate(ctx context.Context, id string) (domain.Endpoint, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Endpoint{}, domain.ErrInvalidInput
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
			ID: uuid.NewString(), Method: src.Method, Path: path,
			Status: src.Status, Response: src.Response, CreatedAt: time.Now().UTC(),
		}
		created, err := s.repo.Create(ctx, dup)
		if errors.Is(err, domain.ErrConflict) {
			continue
		}
		if err != nil {
			return domain.Endpoint{}, err
		}
		return created, nil
	}
	return domain.Endpoint{}, apperror.Conflict("CONFLICT", "could not generate unique duplicate path")
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return domain.ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) Match(ctx context.Context, method, path string) (domain.Endpoint, error) {
	cleanMethod := strings.ToUpper(strings.TrimSpace(method))
	cleanPath := domain.NormalizePath(path)
	m, err := s.repo.FindMatch(ctx, cleanMethod, cleanPath)
	if errors.Is(err, domain.ErrNoMatch) {
		return domain.Endpoint{}, domain.ErrNoMatch
	}
	return m, err
}

func (s *Service) MatchProject(ctx context.Context, projectID, method, path string) (domain.Endpoint, error) {
	cleanMethod := strings.ToUpper(strings.TrimSpace(method))
	cleanPath := domain.NormalizePath(path)
	m, err := s.repo.FindMatchByProject(ctx, projectID, cleanMethod, cleanPath)
	if errors.Is(err, domain.ErrNoMatch) {
		return domain.Endpoint{}, domain.ErrNoMatch
	}
	return m, err
}

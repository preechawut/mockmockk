package mockmock

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"mockapi/pkg/apperr"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if len(p) > 1 {
		p = strings.TrimRight(p, "/")
	}
	return p
}

func (s *Service) List(ctx context.Context) ([]*Mockmock, error) {
	mockmocks, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return mockmocks, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Mockmock, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.NotFound("MOCKMOCK_NOT_FOUND", "mockmock not found")
	}
	return m, nil
}

func (s *Service) Match(ctx context.Context, method, path string) (*Mockmock, error) {
	method = strings.ToUpper(method)
	path = normalizePath(path)
	m, err := s.repo.FindMatch(ctx, method, path)
	if err != nil {
		return nil, apperr.NotFound("NO_MOCKMOCK", "no mockmock configured for "+method+" "+path)
	}
	return m, nil
}

func newMockmock(method, path string, status int, response json.RawMessage) *Mockmock {
	return &Mockmock{
		ID:        uuid.NewString(),
		Method:    method,
		Path:      path,
		Status:    status,
		Response:  response,
		CreatedAt: time.Now().UTC(),
	}
}

func (s *Service) Create(ctx context.Context, in Input) (*Mockmock, error) {
	method := strings.ToUpper(strings.TrimSpace(in.Method))
	if method == "" {
		return nil, apperr.Validation("method is required")
	}
	path := normalizePath(in.Path)
	if path == "/" {
		return nil, apperr.Validation("path is required")
	}
	status := in.Status
	if status == 0 {
		status = 200
	}
	response := in.Response
	if len(response) == 0 {
		response = json.RawMessage("null")
	}
	if existing, _ := s.repo.FindMatch(ctx, method, path); existing != nil {
		return nil, apperr.Conflict("MOCKMOCK_CONFLICT",
			fmt.Sprintf("a mockmock for %s %s already exists", method, path))
	}
	m := newMockmock(method, path, status, response)
	if err := s.repo.Create(ctx, m); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, apperr.Conflict("MOCKMOCK_CONFLICT",
				fmt.Sprintf("a mockmock for %s %s already exists", m.Method, m.Path))
		}
		return nil, apperr.Internal(err)
	}
	return m, nil
}

func (s *Service) Update(ctx context.Context, id string, in Input) (*Mockmock, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.NotFound("MOCKMOCK_NOT_FOUND", "mockmock not found")
	}

	method := strings.ToUpper(strings.TrimSpace(in.Method))
	if method != "" {
		m.Method = method
	}

	if strings.TrimSpace(in.Path) != "" {
		path := normalizePath(in.Path)
		if path == "/" {
			return nil, apperr.Validation("path is required")
		}
		m.Path = path
	}

	if in.Status != 0 {
		m.Status = in.Status
	}
	if len(in.Response) > 0 {
		m.Response = in.Response
	}

	// Check if updating method/path conflicts with another existing mockmock
	if existing, _ := s.repo.FindMatch(ctx, m.Method, m.Path); existing != nil && existing.ID != m.ID {
		return nil, apperr.Conflict("MOCKMOCK_CONFLICT",
			fmt.Sprintf("a mockmock for %s %s already exists", m.Method, m.Path))
	}

	if err := s.repo.Update(ctx, m); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, apperr.Conflict("MOCKMOCK_CONFLICT",
				fmt.Sprintf("a mockmock for %s %s already exists", m.Method, m.Path))
		}
		return nil, apperr.Internal(err)
	}
	return m, nil
}

func (s *Service) Duplicate(ctx context.Context, id string) (*Mockmock, error) {
	src, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.NotFound("MOCKMOCK_NOT_FOUND", "mockmock not found")
	}
	base := src.Path + "-copy"
	path := base
	for i := 2; ; i++ {
		if existing, _ := s.repo.FindMatch(ctx, src.Method, path); existing == nil {
			break
		}
		path = fmt.Sprintf("%s-%d", base, i)
	}
	dup := newMockmock(src.Method, path, src.Status, src.Response)
	if err := s.repo.Create(ctx, dup); err != nil {
		return nil, apperr.Internal(err)
	}
	return dup, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

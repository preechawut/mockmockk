package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"mockapi/internal/modules/project/domain"
)

type Repository interface {
	Create(context.Context, domain.Project) (domain.Project, error)
	List(context.Context) ([]domain.Project, error)
	Get(context.Context, string) (domain.Project, error)
	Rename(context.Context, string, string) (domain.Project, error)
	Delete(context.Context, string) error
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, name, slug string) (domain.Project, error) {
	name = strings.TrimSpace(name)
	if err := domain.Validate(name, slug); err != nil {
		return domain.Project{}, err
	}
	return s.repo.Create(ctx, domain.Project{ID: uuid.NewString(), Name: name, Slug: slug})
}

func (s *Service) List(ctx context.Context) ([]domain.Project, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (domain.Project, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Rename(ctx context.Context, id, name string) (domain.Project, error) {
	name = strings.TrimSpace(name)
	if err := domain.Validate(name, "valid"); err != nil {
		return domain.Project{}, err
	}
	return s.repo.Rename(ctx, id, name)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

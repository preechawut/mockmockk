package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"mockapi/internal/modules/project/domain"
)

type Repository interface {
	Create(ctx context.Context, project domain.Project) (domain.Project, error)
	List(ctx context.Context) ([]domain.Project, error)
	Get(ctx context.Context, id string) (domain.Project, error)
	GetBySlug(ctx context.Context, slug string) (domain.Project, error)
	Rename(ctx context.Context, id, name string) (domain.Project, error)
	Delete(ctx context.Context, id string) error
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, name, slug string) (domain.Project, error) {
	project := domain.Project{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		CreatedAt: time.Now().UTC(),
	}

	if err := project.Validate(); err != nil {
		return domain.Project{}, err
	}

	return s.repository.Create(ctx, project)
}

func (s *Service) List(ctx context.Context) ([]domain.Project, error) {
	return s.repository.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (domain.Project, error) {
	return s.repository.Get(ctx, id)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (domain.Project, error) {
	return s.repository.GetBySlug(ctx, slug)
}

func (s *Service) Rename(ctx context.Context, id, name string) (domain.Project, error) {
	name = strings.TrimSpace(name)
	if err := domain.ValidateName(name); err != nil {
		return domain.Project{}, err
	}

	return s.repository.Rename(ctx, id, name)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}

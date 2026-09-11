package service

import (
	"context"

	"mockapi/internal/modules/endpoint/domain"
)

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

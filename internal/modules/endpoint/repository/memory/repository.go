package memory

import (
	"context"
	"sort"
	"sync"

	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/modules/endpoint/service"
)

type Repository struct {
	mu    sync.RWMutex
	items map[string]domain.Endpoint
}

var _ service.Repository = (*Repository)(nil)

func NewRepository() *Repository {
	return &Repository{
		items: make(map[string]domain.Endpoint),
	}
}

func (r *Repository) Create(_ context.Context, m domain.Endpoint) (domain.Endpoint, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[m.ID]; exists {
		return domain.Endpoint{}, domain.ErrConflict
	}
	for _, item := range r.items {
		if sameProject(item.ProjectID, m.ProjectID) && item.Method == m.Method && item.Path == m.Path {
			return domain.Endpoint{}, domain.ErrConflict
		}
	}

	r.items[m.ID] = clone(m)
	return clone(m), nil
}

func (r *Repository) List(_ context.Context) ([]domain.Endpoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Endpoint, 0, len(r.items))
	for _, item := range r.items {
		result = append(result, clone(item))
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Path == result[j].Path {
			return result[i].Method < result[j].Method
		}
		return result[i].Path < result[j].Path
	})

	return result, nil
}

func (r *Repository) GetByID(_ context.Context, id string) (domain.Endpoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.items[id]
	if !ok {
		return domain.Endpoint{}, domain.ErrNotFound
	}
	return clone(m), nil
}

func (r *Repository) FindMatch(_ context.Context, method, path string) (domain.Endpoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.items {
		if item.Method == method && item.Path == path {
			return clone(item), nil
		}
	}
	return domain.Endpoint{}, domain.ErrNoMatch
}

func (r *Repository) Update(_ context.Context, m domain.Endpoint) (domain.Endpoint, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[m.ID]; !ok {
		return domain.Endpoint{}, domain.ErrNotFound
	}

	for _, item := range r.items {
		if item.ID != m.ID && item.Method == m.Method && item.Path == m.Path {
			return domain.Endpoint{}, domain.ErrConflict
		}
	}

	r.items[m.ID] = clone(m)
	return clone(m), nil
}

func (r *Repository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *Repository) ListByProject(_ context.Context, projectID string) ([]domain.Endpoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Endpoint
	for _, item := range r.items {
		if item.ProjectID != nil && *item.ProjectID == projectID {
			result = append(result, clone(item))
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Path == result[j].Path {
			return result[i].Method < result[j].Method
		}
		return result[i].Path < result[j].Path
	})

	return result, nil
}

func (r *Repository) FindMatchByProject(_ context.Context, projectID, method, path string) (domain.Endpoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.items {
		if item.ProjectID != nil && *item.ProjectID == projectID && item.Method == method && item.Path == path {
			return clone(item), nil
		}
	}
	return domain.Endpoint{}, domain.ErrNoMatch
}

func sameProject(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && b != nil {
		return *a == *b
	}
	return false
}

func clone(m domain.Endpoint) domain.Endpoint {
	respCopy := make([]byte, len(m.Response))
	copy(respCopy, m.Response)
	m.Response = respCopy
	return m
}

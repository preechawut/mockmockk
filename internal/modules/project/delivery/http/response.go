package http

import (
	"time"

	"mockapi/internal/modules/project/domain"
)

type ProjectResponse struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func toProjectResponse(p domain.Project) ProjectResponse {
	return ProjectResponse{
		ID:        p.ID,
		Slug:      p.Slug,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
	}
}

func toProjectResponses(projects []domain.Project) []ProjectResponse {
	responses := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		responses[i] = toProjectResponse(p)
	}
	return responses
}

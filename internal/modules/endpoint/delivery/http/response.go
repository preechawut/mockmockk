package http

import (
	"encoding/json"
	"time"

	"mockapi/internal/modules/endpoint/domain"
)

type EndpointResponse struct {
	ID        string          `json:"id"`
	ProjectID *string         `json:"projectId,omitempty"`
	Method    string          `json:"method"`
	Path      string          `json:"path"`
	Status    int             `json:"status"`
	Response  json.RawMessage `json:"response"`
	CreatedAt time.Time       `json:"createdAt"`
}

func toEndpointResponse(m domain.Endpoint) EndpointResponse {
	resp := m.Response
	if len(resp) == 0 {
		resp = json.RawMessage("null")
	}
	return EndpointResponse{
		ID:        m.ID,
		ProjectID: m.ProjectID,
		Method:    m.Method,
		Path:      m.Path,
		Status:    m.Status,
		Response:  resp,
		CreatedAt: m.CreatedAt,
	}
}

func toEndpointResponses(items []domain.Endpoint) []EndpointResponse {
	res := make([]EndpointResponse, len(items))
	for i, item := range items {
		res[i] = toEndpointResponse(item)
	}
	return res
}

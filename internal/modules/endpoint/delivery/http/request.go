package http

import "encoding/json"

type CreateEndpointRequest struct {
	Method   *string         `json:"method"`
	Path     *string         `json:"path"`
	Status   *int            `json:"status"`
	Response json.RawMessage `json:"response"`
}

type UpdateEndpointRequest struct {
	Method   *string         `json:"method"`
	Path     *string         `json:"path"`
	Status   *int            `json:"status"`
	Response json.RawMessage `json:"response"`
}

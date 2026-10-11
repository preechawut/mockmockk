package http

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/shared/apperror"
)

func (h *Handler) ServeMock(c fiber.Ctx) error {
	reqPath := c.Params("*")
	if reqPath == "" {
		reqPath = "/"
	}
	cleanPath := domain.NormalizePath(reqPath)
	trimmed := strings.TrimPrefix(cleanPath, "/")
	parts := strings.SplitN(trimmed, "/", 2)

	if len(parts) >= 1 && h.projects != nil && parts[0] != "" {
		projectSlug := parts[0]
		project, err := h.projects.GetBySlug(c.Context(), projectSlug)
		if err != nil && !apperror.Is(err, apperror.NotFound) {
			return err
		}
		if err == nil {
			subPath := "/"
			if len(parts) == 2 {
				subPath = "/" + parts[1]
			}
			matched, err := h.endpoints.MatchProject(c.Context(), project.ID, c.Method(), subPath)
			if err != nil {
				return err
			}
			return c.Status(matched.Status).Type("application/json").Send(matched.Response)
		}
	}

	m, err := h.endpoints.Match(c.Context(), c.Method(), cleanPath)
	if err != nil {
		return err
	}
	return c.Status(m.Status).Type("application/json").Send(m.Response)
}

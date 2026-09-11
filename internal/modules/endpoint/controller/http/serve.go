package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"mockapi/internal/modules/endpoint/domain"
)

func (h *Handler) serve(c *fiber.Ctx) error {
	wildcard := c.Params("*")
	cleanPath := domain.NormalizePath(wildcard)

	trimmed := strings.TrimPrefix(cleanPath, "/")
	parts := strings.SplitN(trimmed, "/", 2)

	if len(parts) >= 1 && h.projects != nil && parts[0] != "" {
		projectSlug := parts[0]
		projectID, err := h.projects.Resolve(c.UserContext(), projectSlug, true)
		if err == nil && projectID != "" {
			subPath := "/"
			if len(parts) == 2 {
				subPath = "/" + parts[1]
			}
			endpoint, err := h.endpoints.MatchProject(c.UserContext(), projectID, c.Method(), subPath)
			if err == nil {
				c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
				return c.Status(endpoint.Status).Send(endpoint.Response)
			}
		}
	}

	m, err := h.endpoints.Match(c.UserContext(), c.Method(), cleanPath)
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.Status(m.Status).Send(m.Response)
}

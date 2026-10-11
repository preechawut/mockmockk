package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"mockapi/internal/modules/project/service"
	"mockapi/internal/platform/httpserver"
	"mockapi/internal/shared/apperror"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Create(c fiber.Ctx) error {
	var input CreateProjectRequest
	if err := httpserver.Decode(c, &input); err != nil {
		return err
	}

	result, err := h.service.Create(c.Context(), input.Name, input.Slug)
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderLocation, "/api/v1/projects/"+result.ID)
	return httpserver.JSON(c, fiber.StatusCreated, toProjectResponse(result))
}

func (h *Handler) List(c fiber.Ctx) error {
	projects, err := h.service.List(c.Context())
	if err != nil {
		return err
	}
	return httpserver.JSON(c, fiber.StatusOK, toProjectResponses(projects))
}

func (h *Handler) Get(c fiber.Ctx) error {
	id := c.Params("projectId")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.New(apperror.Invalid, "invalid projectId")
	}

	project, err := h.service.Get(c.Context(), id)
	if err != nil {
		return err
	}
	return httpserver.JSON(c, fiber.StatusOK, toProjectResponse(project))
}

func (h *Handler) Rename(c fiber.Ctx) error {
	id := c.Params("projectId")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.New(apperror.Invalid, "invalid projectId")
	}

	var input RenameProjectRequest
	if err := httpserver.Decode(c, &input); err != nil {
		return err
	}

	project, err := h.service.Rename(c.Context(), id, input.Name)
	if err != nil {
		return err
	}
	return httpserver.JSON(c, fiber.StatusOK, toProjectResponse(project))
}

func (h *Handler) Delete(c fiber.Ctx) error {
	id := c.Params("projectId")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.New(apperror.Invalid, "invalid projectId")
	}

	if err := h.service.Delete(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

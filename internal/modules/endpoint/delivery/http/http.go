package http

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"mockapi/internal/modules/endpoint/service"
	projectdomain "mockapi/internal/modules/project/domain"
	"mockapi/internal/platform/httpserver"
	"mockapi/internal/shared/apperror"
)

type ProjectService interface {
	GetBySlug(ctx context.Context, slug string) (projectdomain.Project, error)
}

type Handler struct {
	endpoints *service.Service
	projects  ProjectService
}

func New(endpoints *service.Service, projects ProjectService) *Handler {
	return &Handler{
		endpoints: endpoints,
		projects:  projects,
	}
}

// Global endpoints: /api/v1/mockmocks

func (h *Handler) Create(c fiber.Ctx) error {
	var body CreateEndpointRequest
	if err := httpserver.Decode(c, &body); err != nil {
		return err
	}
	if body.Method == nil || body.Path == nil {
		return apperror.New(apperror.Invalid, "method and path are required")
	}
	status := 0
	if body.Status != nil {
		status = *body.Status
	}

	created, err := h.endpoints.Create(c.Context(), *body.Method, *body.Path, status, body.Response)
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderLocation, "/api/v1/mockmocks/"+created.ID)
	return httpserver.JSON(c, fiber.StatusCreated, toEndpointResponse(created))
}

func (h *Handler) List(c fiber.Ctx) error {
	items, err := h.endpoints.List(c.Context())
	if err != nil {
		return err
	}
	return httpserver.JSON(c, fiber.StatusOK, toEndpointResponses(items))
}

func (h *Handler) Get(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.New(apperror.Invalid, "invalid id")
	}

	item, err := h.endpoints.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	return httpserver.JSON(c, fiber.StatusOK, toEndpointResponse(item))
}

func (h *Handler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.New(apperror.Invalid, "invalid id")
	}

	var body UpdateEndpointRequest
	if err := httpserver.Decode(c, &body); err != nil {
		return err
	}

	method := ""
	if body.Method != nil {
		method = *body.Method
	}
	path := ""
	if body.Path != nil {
		path = *body.Path
	}
	status := 0
	if body.Status != nil {
		status = *body.Status
	}

	updated, err := h.endpoints.Update(c.Context(), id, method, path, status, body.Response)
	if err != nil {
		return err
	}
	return httpserver.JSON(c, fiber.StatusOK, toEndpointResponse(updated))
}

func (h *Handler) Duplicate(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.New(apperror.Invalid, "invalid id")
	}

	dup, err := h.endpoints.Duplicate(c.Context(), id)
	if err != nil {
		return err
	}
	location := "/api/v1/mockmocks/" + dup.ID
	if dup.ProjectID != nil {
		location = "/api/v1/projects/" + *dup.ProjectID + "/endpoints/" + dup.ID
	}
	c.Set(fiber.HeaderLocation, location)
	return httpserver.JSON(c, fiber.StatusCreated, toEndpointResponse(dup))
}

func (h *Handler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.New(apperror.Invalid, "invalid id")
	}

	if err := h.endpoints.Delete(c.Context(), id); err != nil {
		return err
	}
	return httpserver.JSON(c, fiber.StatusOK, fiber.Map{"deleted": true})
}

// Project endpoints: /api/v1/projects/:projectId/endpoints

func (h *Handler) CreateProjectEndpoint(c fiber.Ctx) error {
	projectID := c.Params("projectId")
	if _, err := uuid.Parse(projectID); err != nil {
		return apperror.New(apperror.Invalid, "invalid projectId")
	}

	var body CreateEndpointRequest
	if err := httpserver.Decode(c, &body); err != nil {
		return err
	}
	if body.Method == nil || body.Path == nil {
		return apperror.New(apperror.Invalid, "method and path are required")
	}
	status := 0
	if body.Status != nil {
		status = *body.Status
	}

	created, err := h.endpoints.CreateWithProject(c.Context(), &projectID, *body.Method, *body.Path, status, body.Response)
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderLocation, "/api/v1/projects/"+projectID+"/endpoints/"+created.ID)
	return httpserver.JSON(c, fiber.StatusCreated, toEndpointResponse(created))
}

func (h *Handler) ListProjectEndpoints(c fiber.Ctx) error {
	projectID := c.Params("projectId")
	if _, err := uuid.Parse(projectID); err != nil {
		return apperror.New(apperror.Invalid, "invalid projectId")
	}

	items, err := h.endpoints.ListByProject(c.Context(), projectID)
	if err != nil {
		return err
	}
	return httpserver.JSON(c, fiber.StatusOK, toEndpointResponses(items))
}

func (h *Handler) GetProjectEndpoint(c fiber.Ctx) error {
	if err := h.validateProjectOwnership(c); err != nil {
		return err
	}
	return h.Get(c)
}

func (h *Handler) UpdateProjectEndpoint(c fiber.Ctx) error {
	if err := h.validateProjectOwnership(c); err != nil {
		return err
	}
	return h.Update(c)
}

func (h *Handler) DuplicateProjectEndpoint(c fiber.Ctx) error {
	if err := h.validateProjectOwnership(c); err != nil {
		return err
	}
	return h.Duplicate(c)
}

func (h *Handler) DeleteProjectEndpoint(c fiber.Ctx) error {
	if err := h.validateProjectOwnership(c); err != nil {
		return err
	}
	return h.Delete(c)
}

func (h *Handler) validateProjectOwnership(c fiber.Ctx) error {
	projectID := c.Params("projectId")
	if _, err := uuid.Parse(projectID); err != nil {
		return apperror.New(apperror.Invalid, "invalid projectId")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.New(apperror.Invalid, "invalid id")
	}

	item, err := h.endpoints.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	if item.ProjectID == nil || *item.ProjectID != projectID {
		return apperror.New(apperror.NotFound, "endpoint not found in this project")
	}
	return nil
}

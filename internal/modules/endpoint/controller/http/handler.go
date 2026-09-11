package http

import (
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	httpkit "mockapi/internal/controller/http"
	"mockapi/internal/modules/endpoint/domain"
)

type ProjectLookup interface {
	Resolve(ctx context.Context, ref string, slug bool) (string, error)
}

type EndpointService interface {
	List(ctx context.Context) ([]domain.Endpoint, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.Endpoint, error)
	GetByID(ctx context.Context, id string) (domain.Endpoint, error)
	Match(ctx context.Context, method, path string) (domain.Endpoint, error)
	MatchProject(ctx context.Context, projectID, method, path string) (domain.Endpoint, error)
	Create(ctx context.Context, method, path string, status int, response json.RawMessage) (domain.Endpoint, error)
	CreateWithProject(ctx context.Context, projectID *string, method, path string, status int, response json.RawMessage) (domain.Endpoint, error)
	Update(ctx context.Context, id, method, path string, status int, response json.RawMessage) (domain.Endpoint, error)
	Duplicate(ctx context.Context, id string) (domain.Endpoint, error)
	Delete(ctx context.Context, id string) error
}

type Handler struct {
	endpoints EndpointService
	projects  ProjectLookup
}

func NewHandler(endpoints EndpointService, projects ...ProjectLookup) *Handler {
	var pl ProjectLookup
	if len(projects) > 0 {
		pl = projects[0]
	}
	return &Handler{endpoints: endpoints, projects: pl}
}

// routes
func (h *Handler) RegisterRoutes(r fiber.Router) {
	group := r.Group("/mockmocks")
	group.Post("/", h.create)
	group.Get("/", h.list)
	group.Get("/:id", h.get)
	group.Put("/:id", h.update)
	group.Patch("/:id", h.update)
	group.Post("/:id/duplicate", h.duplicate)
	group.Delete("/:id", h.delete)

	projectsGroup := r.Group("/projects/:projectId/endpoints")
	projectsGroup.Post("/", h.createProjectEndpoint)
	projectsGroup.Get("/", h.listProjectEndpoints)
	projectsGroup.Get("/:id", h.get)
	projectsGroup.Put("/:id", h.update)
	projectsGroup.Patch("/:id", h.update)
	projectsGroup.Post("/:id/duplicate", h.duplicate)
	projectsGroup.Delete("/:id", h.delete)
}

func (h *Handler) RegisterRootRoutes(r fiber.Router) {
	r.All("/mock", h.serve)
	r.All("/mock/*", h.serve)
}

func (h *Handler) createProjectEndpoint(c *fiber.Ctx) error {
	projectID, err := httpkit.PathUUID(c, "projectId")
	if err != nil {
		return err
	}
	var body CreateEndpointRequest
	if err := httpkit.DecodeJSON(c, &body); err != nil {
		return err
	}
	if body.Method == nil || body.Path == nil {
		return httpkit.ErrMissingField
	}
	status := 0
	if body.Status != nil {
		status = *body.Status
	}
	var resp json.RawMessage
	if body.Response != nil {
		resp = *body.Response
	}

	created, err := h.endpoints.CreateWithProject(c.UserContext(), &projectID, *body.Method, *body.Path, status, resp)
	if err != nil {
		return err
	}
	c.Set("Location", "/api/v1/projects/"+projectID+"/endpoints/"+created.ID)
	return c.Status(fiber.StatusCreated).JSON(ToEndpointResponse(created))
}

func (h *Handler) listProjectEndpoints(c *fiber.Ctx) error {
	projectID, err := httpkit.PathUUID(c, "projectId")
	if err != nil {
		return err
	}
	items, err := h.endpoints.ListByProject(c.UserContext(), projectID)
	if err != nil {
		return err
	}
	return c.JSON(ToEndpointResponses(items))
}

func (h *Handler) create(c *fiber.Ctx) error {
	var body CreateEndpointRequest
	if err := httpkit.DecodeJSON(c, &body); err != nil {
		return err
	}
	if body.Method == nil || body.Path == nil {
		return httpkit.ErrMissingField
	}
	status := 0
	if body.Status != nil {
		status = *body.Status
	}
	var resp json.RawMessage
	if body.Response != nil {
		resp = *body.Response
	}

	created, err := h.endpoints.Create(c.UserContext(), *body.Method, *body.Path, status, resp)
	if err != nil {
		return err
	}
	c.Set("Location", "/api/v1/mockmocks/"+created.ID)
	return c.Status(fiber.StatusCreated).JSON(ToEndpointResponse(created))
}

func (h *Handler) list(c *fiber.Ctx) error {
	items, err := h.endpoints.List(c.UserContext())
	if err != nil {
		return err
	}
	return c.JSON(ToEndpointResponses(items))
}

func (h *Handler) get(c *fiber.Ctx) error {
	id, err := httpkit.PathUUID(c, "id")
	if err != nil {
		return err
	}
	item, err := h.endpoints.GetByID(c.UserContext(), id)
	if err != nil {
		return err
	}
	return c.JSON(ToEndpointResponse(item))
}

func (h *Handler) update(c *fiber.Ctx) error {
	id, err := httpkit.PathUUID(c, "id")
	if err != nil {
		return err
	}
	var body UpdateEndpointRequest
	if err := httpkit.DecodeJSON(c, &body); err != nil {
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
	var resp json.RawMessage
	if body.Response != nil {
		resp = *body.Response
	}

	updated, err := h.endpoints.Update(c.UserContext(), id, method, path, status, resp)
	if err != nil {
		return err
	}
	return c.JSON(ToEndpointResponse(updated))
}

func (h *Handler) duplicate(c *fiber.Ctx) error {
	id, err := httpkit.PathUUID(c, "id")
	if err != nil {
		return err
	}
	dup, err := h.endpoints.Duplicate(c.UserContext(), id)
	if err != nil {
		return err
	}
	c.Set("Location", "/api/v1/mockmocks/"+dup.ID)
	return c.Status(fiber.StatusCreated).JSON(ToEndpointResponse(dup))
}

func (h *Handler) delete(c *fiber.Ctx) error {
	id, err := httpkit.PathUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.endpoints.Delete(c.UserContext(), id); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"deleted": true})
}

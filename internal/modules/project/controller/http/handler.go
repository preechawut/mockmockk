package http

import (
	"github.com/gofiber/fiber/v2"
	httpkit "mockapi/internal/controller/http"
	"mockapi/internal/modules/project/service"
)

type Handler struct{ service *service.Service }

func NewHandler(s *service.Service) *Handler { return &Handler{service: s} }
func (h *Handler) RegisterRoutes(r fiber.Router) {
	g := r.Group("/projects")
	g.Post("/", h.create)
	g.Get("/", h.list)
	g.Get("/:projectId", h.get)
	g.Patch("/:projectId", h.rename)
	g.Delete("/:projectId", h.delete)
}
func (h *Handler) create(c *fiber.Ctx) error {
	var b struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := httpkit.DecodeJSON(c, &b); err != nil {
		return err
	}
	p, err := h.service.Create(c.UserContext(), b.Name, b.Slug)
	if err != nil {
		return err
	}
	c.Set("Location", "/api/v1/projects/"+p.ID)
	return c.Status(201).JSON(p)
}
func (h *Handler) list(c *fiber.Ctx) error {
	p, err := h.service.List(c.UserContext())
	if err != nil {
		return err
	}
	return c.JSON(p)
}
func (h *Handler) get(c *fiber.Ctx) error {
	id, err := httpkit.PathUUID(c, "projectId")
	if err != nil {
		return err
	}
	p, err := h.service.Get(c.UserContext(), id)
	if err != nil {
		return err
	}
	return c.JSON(p)
}
func (h *Handler) rename(c *fiber.Ctx) error {
	id, err := httpkit.PathUUID(c, "projectId")
	if err != nil {
		return err
	}
	var b struct {
		Name string `json:"name"`
	}
	if err := httpkit.DecodeJSON(c, &b); err != nil {
		return err
	}
	p, err := h.service.Rename(c.UserContext(), id, b.Name)
	if err != nil {
		return err
	}
	return c.JSON(p)
}
func (h *Handler) delete(c *fiber.Ctx) error {
	id, err := httpkit.PathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if err := h.service.Delete(c.UserContext(), id); err != nil {
		return err
	}
	return c.SendStatus(204)
}

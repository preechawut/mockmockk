package app

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v3"

	"mockapi/internal/config"
	endpointdelivery "mockapi/internal/modules/endpoint/delivery/http"
	projectdelivery "mockapi/internal/modules/project/delivery/http"
	"mockapi/internal/platform/httpserver"
	"mockapi/internal/platform/httpserver/middleware"
)

func newRouter(db *sql.DB, cfg config.Config, project *projectdelivery.Handler, endpoint *endpointdelivery.Handler) *fiber.App {
	router := httpserver.New(cfg.TrustedProxies)

	// middleware
	router.Use(httpserver.ResponseHeaders())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logging())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	router.Use(middleware.RateLimit(cfg.APIRateLimit, time.Minute))
	router.Use(httpserver.RequestTimeout(15 * time.Second))
	router.Use(middleware.Recovery())

	// health checks / probe
	router.Get("/healthz", func(c fiber.Ctx) error {
		return httpserver.JSON(c, fiber.StatusOK, fiber.Map{"status": "ok"})
	})

	router.Get("/readyz", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			return httpserver.ErrorStatus(c, fiber.StatusServiceUnavailable, "database_unavailable", "database is unavailable")
		}

		return httpserver.JSON(c, fiber.StatusOK, fiber.Map{"status": "ready"})
	})

	// api v1
	v1 := router.Group("/api/v1")

	// projects
	projects := v1.Group("/projects")
	projects.Post("", project.Create)
	projects.Get("", project.List)
	projects.Get("/:projectId", project.Get)
	projects.Patch("/:projectId", project.Rename)
	projects.Delete("/:projectId", project.Delete)

	// project endpoints
	projectEndpoints := projects.Group("/:projectId/endpoints")
	projectEndpoints.Post("", endpoint.CreateProjectEndpoint)
	projectEndpoints.Get("", endpoint.ListProjectEndpoints)
	projectEndpoints.Get("/:id", endpoint.GetProjectEndpoint)
	projectEndpoints.Put("/:id", endpoint.UpdateProjectEndpoint)
	projectEndpoints.Patch("/:id", endpoint.UpdateProjectEndpoint)
	projectEndpoints.Post("/:id/duplicate", endpoint.DuplicateProjectEndpoint)
	projectEndpoints.Delete("/:id", endpoint.DeleteProjectEndpoint)

	// global endpoints
	mockmocks := v1.Group("/mockmocks")
	mockmocks.Post("", endpoint.Create)
	mockmocks.Get("", endpoint.List)
	mockmocks.Get("/:id", endpoint.Get)
	mockmocks.Put("/:id", endpoint.Update)
	mockmocks.Patch("/:id", endpoint.Update)
	mockmocks.Post("/:id/duplicate", endpoint.Duplicate)
	mockmocks.Delete("/:id", endpoint.Delete)

	// mock server route
	router.All("/mock", endpoint.ServeMock)
	router.All("/mock/*", endpoint.ServeMock)

	return router
}

package app

import (
	"database/sql"

	"github.com/gofiber/fiber/v3"

	"mockapi/internal/config"
	endpointdelivery "mockapi/internal/modules/endpoint/delivery/http"
	endpointrepo "mockapi/internal/modules/endpoint/infrastructure"
	endpointservice "mockapi/internal/modules/endpoint/service"
	projectdelivery "mockapi/internal/modules/project/delivery/http"
	projectrepo "mockapi/internal/modules/project/infrastructure"
	projectservice "mockapi/internal/modules/project/service"
)

func New(db *sql.DB, cfg config.Config) *fiber.App {
	projectRepository := projectrepo.New(db)
	projectService := projectservice.New(projectRepository)
	projectHandler := projectdelivery.New(projectService)

	endpointRepository := endpointrepo.New(db)
	endpointService := endpointservice.New(endpointRepository)
	endpointHandler := endpointdelivery.New(endpointService, projectService)

	return newRouter(db, cfg, projectHandler, endpointHandler)
}

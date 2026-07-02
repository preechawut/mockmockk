package server

import (
	"context"
	"net/http"
	"time"

	"mockapi/internal/mockmock"
	"mockapi/pkg/httputil"
	"mockapi/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(mockmockH *mockmock.Handler, pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(middleware.CORS(), gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			httputil.RespondAPIError(c, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database is not reachable")
			return
		}
		httputil.RespondOK(c, gin.H{"status": "ok", "db": "connected"})
	})

	mockmockH.RegisterRoutes(r)

	return r
}

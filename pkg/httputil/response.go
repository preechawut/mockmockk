package httputil

import (
	"errors"
	"log"
	"net/http"

	"mockapi/pkg/apperr"

	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RespondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

func RespondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}

func RespondAPIError(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorResponse{
		Code:    code,
		Message: message,
	})
}

func RespondErr(c *gin.Context, err error) {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		if ae.Status >= http.StatusInternalServerError {
			logErr(c, err)
		}
		RespondAPIError(c, ae.Status, ae.Code, ae.Message)
		return
	}
	logErr(c, err)
	RespondAPIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
}

func logErr(c *gin.Context, err error) {
	log.Printf("%s %s: %v", c.Request.Method, c.Request.URL.Path, err)
}

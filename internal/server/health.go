package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (s *Server) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{
		Status:  "healthy",
		Message: "Service is running",
	})
}

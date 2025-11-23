package server

import (
	"github.com/vnchk1/pr-manager/internal/middleware"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/vnchk1/pr-manager/internal/service"

	"github.com/labstack/echo/v4"
)

type Server struct {
	echo    *echo.Echo
	port    int
	service *service.Service
}

func NewServer(port int, service *service.Service, logger *slog.Logger) *Server {
	e := echo.New()

	e.Use(middleware.LoggingMiddleware(logger))

	server := &Server{
		echo:    e,
		port:    port,
		service: service,
	}

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	s.echo.GET("/health", s.healthCheck)

	s.echo.POST("/team/add", s.createTeam)
	s.echo.GET("/team/get", s.getTeam)

	s.echo.POST("/users/setIsActive", s.setUserActive)
	s.echo.GET("/users/getReview", s.getUserReviewPRs)

	s.echo.POST("/pullRequest/create", s.createPR)
	s.echo.POST("/pullRequest/merge", s.mergePR)
	s.echo.POST("/pullRequest/reassign", s.reassignReviewer)

	s.echo.GET("/stats/assignments", s.getStats)
	s.echo.GET("/stats/user", s.getUserStats)
}

func (s *Server) Start(logger *slog.Logger) error {
	address := fmt.Sprintf(":%d", s.port)
	logger.Debug(fmt.Sprintf("Starting server on %s", address))
	return s.echo.Start(address)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

func (s *Server) GracefulStart(logger *slog.Logger) error {

	go func() {
		if err := s.Start(logger); err != nil && err != http.ErrServerClosed {
			logger.Debug("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	BaseResponse
	Error string `json:"error,omitempty"`
}

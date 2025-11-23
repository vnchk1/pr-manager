package main

import (
	"github.com/vnchk1/pr-manager/internal/config"
	"github.com/vnchk1/pr-manager/internal/db"
	logpkg "github.com/vnchk1/pr-manager/internal/logger"
	"github.com/vnchk1/pr-manager/internal/migration"
	"github.com/vnchk1/pr-manager/internal/server"
	"github.com/vnchk1/pr-manager/internal/service"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logpkg.NewLogger(cfg.LogLevel)

	postgres, err := db.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	logger.Debug("Connected to database")
	defer postgres.Close()

	if err = migration.RunMigrations(cfg.Database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	logger.Debug("Migrations run")

	services := service.New(postgres.Repo)
	logger.Debug("Services initialized successfully")

	srv := server.NewServer(cfg.AppPort, services, logger)

	if err = srv.GracefulStart(logger); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server exited")
}

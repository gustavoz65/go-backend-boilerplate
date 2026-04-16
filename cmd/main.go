package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/go-backend-boilerplate/internal/config"
	"github.com/example/go-backend-boilerplate/internal/database"
	"github.com/example/go-backend-boilerplate/internal/logger"
	"github.com/example/go-backend-boilerplate/internal/router"
	"github.com/example/go-backend-boilerplate/internal/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	loggerService := logger.NewLoggerService(cfg.Observability)
	defer loggerService.Shutdown()

	log := logger.NewLoggerWithService(cfg.Observability, loggerService)

	if cfg.Primary.Env != "local" {
		if err := database.Migrate(context.Background(), &log, cfg); err != nil {
			log.Fatal().Err(err).Msg("failed to migrate database")
		}
	}

	srv, err := server.New(cfg, &log, loggerService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize server")
	}

	e := router.New(cfg, srv.DB, &log, srv)
	srv.SetHandler(e)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			log.Info().Err(err).Msg("server stopped")
		}
	}()

	log.Info().Msgf("server started on port %s", cfg.Server.Port)

	<-quit
	log.Info().Msg("received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("forced shutdown")
	}
}

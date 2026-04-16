package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	zerolog "github.com/rs/zerolog"

	"github.com/example/go-backend-boilerplate/internal/config"
	"github.com/example/go-backend-boilerplate/internal/database"
	"github.com/example/go-backend-boilerplate/internal/lib/firebase"
	"github.com/example/go-backend-boilerplate/internal/lib/utils/job"
	"github.com/example/go-backend-boilerplate/internal/logger"
)

type Server struct {
	Config         *config.Config
	Logger         *zerolog.Logger
	LoggerService  *logger.LoggerService
	DB             *database.Database
	Redis          *redis.Client
	FirebaseClient *firebase.Client
	httpServer     *http.Server
	Job            *job.JobService
}

func New(cfg *config.Config, log *zerolog.Logger, loggerService *logger.LoggerService) (*Server, error) {
	db, err := database.New(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to connect to Redis, continuing without Redis")
	} else {
		log.Info().Msg("connected to Redis")
	}

	// Initialize Firebase
	var firebaseClient *firebase.Client
	if cfg.Firebase.Enabled {
		firebaseClient, err = firebase.NewClient(&cfg.Firebase)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize Firebase, continuing without Firebase")
		} else {
			log.Info().Msg("Firebase initialized successfully")
		}
	}

	// job service
	jobService := job.NewJobService(log, cfg)
	jobService.InitHandlers(cfg, log)

	// Start job server
	if err := jobService.Start(); err != nil {
		return nil, err
	}

	server := &Server{
		Config:         cfg,
		Logger:         log,
		LoggerService:  loggerService,
		DB:             db,
		Redis:          redisClient,
		FirebaseClient: firebaseClient,
		Job:            jobService,
	}

	return server, nil
}

func (s *Server) Start() error {
	var handler http.Handler
	if s.httpServer != nil {
		handler = s.httpServer.Handler
	}

	s.httpServer = &http.Server{
		Addr:         ":" + s.Config.Server.Port,
		Handler:      handler,
		ReadTimeout:  time.Duration(s.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.Config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.Config.Server.IdleTimeout) * time.Second,
	}

	if s.Config.Server.TLSEnabled {
		if s.Config.Server.TLSCertFile == "" || s.Config.Server.TLSKeyFile == "" {
			return fmt.Errorf("TLS habilitado mas cert_file ou key_file não configurados")
		}
		s.Logger.Info().Msgf("starting HTTPS server on port %s", s.Config.Server.Port)
		return s.httpServer.ListenAndServeTLS(s.Config.Server.TLSCertFile, s.Config.Server.TLSKeyFile)
	}

	s.Logger.Info().Msgf("starting HTTP server on port %s", s.Config.Server.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.Logger.Info().Msg("shutting down server...")

	// Shutdown HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.Logger.Error().Err(err).Msg("error shutting down HTTP server")
		}
	}

	// Stop job service
	if s.Job != nil {
		s.Job.Stop()
	}

	// Close Redis connection
	if s.Redis != nil {
		if err := s.Redis.Close(); err != nil {
			s.Logger.Error().Err(err).Msg("error closing Redis connection")
		}
	}

	// Close database connection
	if s.DB != nil {
		if err := s.DB.Close(); err != nil {
			s.Logger.Error().Err(err).Msg("error closing database connection")
		}
	}

	s.Logger.Info().Msg("server shutdown complete")
	return nil
}

func (s *Server) SetHandler(handler http.Handler) {
	if s.httpServer == nil {
		s.httpServer = &http.Server{}
	}
	s.httpServer.Handler = handler
}

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/config"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/database"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/handler"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/middleware"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/modules/auth"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/modules/notifications"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/modules/users"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/repository"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/server"
)

// New cria e configura o router Gin, compondo as rotas registradas por cada
// módulo de domínio (internal/modules/*). O router principal não conhece
// os detalhes de cada domínio — apenas monta os middlewares compartilhados
// e chama o RegisterRoutes de cada módulo.
func New(cfg *config.Config, db *database.Database, logger *zerolog.Logger, srv *server.Server) *gin.Engine {
	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware(cfg.Server.CORSAllowedOrigins))
	r.Use(middleware.CSRFTokenGenerator())
	r.Use(middleware.ErrorHandler(logger))

	// Repositórios
	userRepo := repository.NewUserRepository(db, logger)
	providerRepo := repository.NewOAuthProviderRepository(db, logger)
	notificationRepo := repository.NewNotificationRepository(db, logger)
	auditLogRepo := repository.NewAuditLogRepository(db, logger)

	// Services + handlers de cada módulo
	authService := auth.NewService(userRepo, providerRepo, srv.FirebaseClient, cfg, logger)
	userService := users.NewService(userRepo, logger)
	notificationService := notifications.NewService(notificationRepo, logger)

	authHandler := auth.NewHandler(authService)
	userHandler := users.NewHandler(userService)
	notificationHandler := notifications.NewHandler(notificationService)
	healthHandler := handler.NewHealthHandler(srv)

	r.GET("/health", healthHandler.CheckHandler)

	api := r.Group("/api/v1")

	csrfTokenGen := middleware.CSRFTokenGenerator()
	api.GET("/csrf-token", csrfTokenGen, func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]string{"csrf_token": c.Writer.Header().Get("X-CSRF-Token")})
	})

	// Middlewares compartilhados entre os módulos
	requireAuth := auth.RequireAuth(authService)
	auditMiddleware := middleware.NewAuditMiddleware(auditLogRepo, logger)
	csrfMiddleware := middleware.CSRFMiddleware()
	mutationRL := middleware.MutationRateLimit(srv.Redis)
	readRL := middleware.ReadRateLimit(srv.Redis)

	auth.RegisterRoutes(api, authHandler, auth.Middlewares{
		AuthRateLimit:    middleware.AuthRateLimit(srv.Redis),
		RefreshRateLimit: middleware.RefreshRateLimit(srv.Redis),
		CSRFTokenGen:     csrfTokenGen,
		RequireAuth:      requireAuth,
		Audit:            auditMiddleware.Handler(),
		CSRF:             csrfMiddleware,
		MutationRL:       mutationRL,
		ReadRL:           readRL,
	})

	users.RegisterRoutes(api, userHandler, users.Middlewares{
		RequireAuth: requireAuth,
		Audit:       auditMiddleware.Handler(),
		CSRF:        csrfMiddleware,
		MutationRL:  mutationRL,
		ReadRL:      readRL,
	})

	notifications.RegisterRoutes(api, notificationHandler, notifications.Middlewares{
		RequireAuth: requireAuth,
		Audit:       auditMiddleware.Handler(),
		CSRF:        csrfMiddleware,
		MutationRL:  mutationRL,
		ReadRL:      readRL,
	})

	return r
}

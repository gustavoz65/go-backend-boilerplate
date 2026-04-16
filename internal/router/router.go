package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"

	"github.com/example/go-backend-boilerplate/internal/config"
	"github.com/example/go-backend-boilerplate/internal/database"
	"github.com/example/go-backend-boilerplate/internal/handler"
	"github.com/example/go-backend-boilerplate/internal/lib/utils/validator"
	"github.com/example/go-backend-boilerplate/internal/middleware"
	"github.com/example/go-backend-boilerplate/internal/repository"
	"github.com/example/go-backend-boilerplate/internal/server"
	"github.com/example/go-backend-boilerplate/internal/service"
)

// New creates and configures the Echo router for the core boilerplate modules.
func New(cfg *config.Config, db *database.Database, logger *zerolog.Logger, srv *server.Server) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = validator.New()

	e.HTTPErrorHandler = middleware.ErrorHandler(logger)
	e.Use(echomiddleware.RequestID())
	e.Use(middleware.RecoveryMiddleware(logger))
	e.Use(middleware.LoggerMiddleware(logger))
	e.Use(middleware.SecurityHeadersMiddleware())
	e.Use(middleware.CORSMiddleware(cfg.Server.CORSAllowedOrigins))
	e.Use(middleware.CSRFTokenGenerator())

	userRepo := repository.NewUserRepository(db, logger)
	providerRepo := repository.NewOAuthProviderRepository(db, logger)
	notificationRepo := repository.NewNotificationRepository(db, logger)
	auditLogRepo := repository.NewAuditLogRepository(db, logger)

	authService := service.NewAuthService(userRepo, providerRepo, srv.FirebaseClient, cfg, logger)
	userService := service.NewUserService(userRepo, logger)
	notificationService := service.NewNotificationService(notificationRepo, logger)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	healthHandler := handler.NewHealthHandler(srv)

	e.GET("/health", healthHandler.CheckHandler)

	api := e.Group("/api/v1")
	authRateLimiter := middleware.AuthRateLimit(srv.Redis)
	refreshRateLimiter := middleware.RefreshRateLimit(srv.Redis)
	csrfTokenGen := middleware.CSRFTokenGenerator()

	api.GET("/csrf-token", func(c echo.Context) error {
		token := c.Response().Header().Get("X-CSRF-Token")
		return c.JSON(http.StatusOK, map[string]string{"csrf_token": token})
	}, csrfTokenGen)

	authWithRL := api.Group("/auth", authRateLimiter, csrfTokenGen)
	authRefresh := api.Group("/auth", refreshRateLimiter, csrfTokenGen)

	authWithRL.POST("/register", authHandler.Register)
	authWithRL.POST("/login", authHandler.Login)
	authWithRL.POST("/social/login", authHandler.SocialLogin)
	authRefresh.POST("/refresh", authHandler.RefreshToken)
	authRefresh.POST("/logout", authHandler.Logout)

	authMiddleware := middleware.AuthMiddleware(authService)
	auditMiddleware := middleware.NewAuditMiddleware(auditLogRepo, logger)
	csrfMiddleware := middleware.CSRFMiddleware()
	mutationRL := middleware.MutationRateLimit(srv.Redis)
	readRL := middleware.ReadRateLimit(srv.Redis)

	authProtected := api.Group("/auth", authMiddleware, auditMiddleware.Handler(), csrfMiddleware)
	authProtected.POST("/change-password", authHandler.ChangePassword, mutationRL)
	authProtected.POST("/set-password", authHandler.SetPassword, mutationRL)
	authProtected.POST("/social/link", authHandler.LinkProvider, mutationRL)
	authProtected.DELETE("/social/:provider", authHandler.UnlinkProvider, mutationRL)
	authProtected.GET("/social/providers", authHandler.GetLinkedProviders, readRL)

	users := api.Group("/users", authMiddleware, auditMiddleware.Handler(), csrfMiddleware)
	users.GET("/me", userHandler.GetMe, readRL)
	users.PUT("/me", userHandler.UpdateMe, mutationRL)
	users.DELETE("/me", userHandler.DeactivateMe, mutationRL)
	users.GET("/settings", userHandler.GetSettings, readRL)
	users.PUT("/settings", userHandler.UpdateSettings, mutationRL)
	users.PATCH("/me/onboarding-complete", userHandler.CompleteOnboarding, mutationRL)

	notifications := api.Group("/notifications", authMiddleware, auditMiddleware.Handler(), csrfMiddleware)
	notifications.GET("", notificationHandler.GetAll, readRL)
	notifications.GET("/unread", notificationHandler.GetUnread, readRL)
	notifications.GET("/unread/count", notificationHandler.GetUnreadCount, readRL)
	notifications.PATCH("/:id/read", notificationHandler.MarkAsRead, mutationRL)
	notifications.PATCH("/read-all", notificationHandler.MarkAllAsRead, mutationRL)
	notifications.DELETE("/:id", notificationHandler.Delete, mutationRL)

	return e
}

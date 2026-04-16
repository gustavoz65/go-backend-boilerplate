package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/middleware"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/server"
)

type HealthHandler struct {
	Handler
}

func NewHealthHandler(s *server.Server) *HealthHandler {
	return &HealthHandler{
		Handler: NewHandler(s),
	}
}

func (h *HealthHandler) CheckHandler(c echo.Context) error {
	start := time.Now()
	logger := middleware.GetLogger(c).With().
		Str("operation", "health_check").
		Logger()

	response := map[string]interface{}{
		"status":      "ok",
		"timestamp":   time.Now().UTC(),
		"environment": h.server.Config.Primary.Env,
		"checks":      make(map[string]interface{}),
	}

	checks := response["checks"].(map[string]interface{})
	isHealthy := true

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbStart := time.Now()
	if err := h.server.DB.Pool.Ping(); err != nil {
		checks["database"] = map[string]interface{}{
			"status":        "error",
			"response_time": time.Since(dbStart).String(),
			"error":         err.Error(),
		}
		isHealthy = false
		logger.Error().Err(err).Dur("response_time", time.Since(dbStart)).Msg("o Banco de Dados não está Saudavel")
		if h.server.LoggerService != nil && h.server.LoggerService.GetApplication() != nil {
			h.server.LoggerService.GetApplication().RecordCustomEvent(
				"HealthCheckError", map[string]interface{}{
					"check_type":    "database",
					"operation":     "health_check",
					"error_type":    "database_unhealthy",
					"error_message": err.Error(),
				})
		}
	} else {
		checks["database"] = map[string]interface{}{
			"status":        "healthy",
			"response_time": time.Since(dbStart).String(),
		}
		logger.Info().Dur("response_time", time.Since(dbStart)).Msg("O Banco de Dados está Saudavel")
	}

	if h.server.Redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		redisStart := time.Now()
		if err := h.server.Redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = map[string]interface{}{
				"status":        "unhealthy",
				"response_time": time.Since(redisStart).String(),
				"error":         err.Error(),
			}
			logger.Error().Err(err).Dur("response_time", time.Since(redisStart)).Msg("O Redis não está Saudavel")
			if h.server.LoggerService != nil && h.server.LoggerService.GetApplication() != nil {
				h.server.LoggerService.GetApplication().RecordCustomEvent(
					"HealthCheckError", map[string]interface{}{
						"check_type":    "redis",
						"operation":     "health_check",
						"error_type":    "redis_unhealthy",
						"error_message": err.Error(),
					})
			}
		} else {
			checks["redis"] = map[string]interface{}{
				"status":        "healthy",
				"response_time": time.Since(redisStart).String(),
			}
			logger.Info().Dur("response_time", time.Since(redisStart)).Msg("O Redis está Saudavel")
		}
	}

	if !isHealthy {
		response["status"] = "unhealthy"
		logger.Warn().
			Dur("total_duration", time.Since(start)).
			Msg("O health check falhou")
		if h.server.LoggerService != nil && h.server.LoggerService.GetApplication() != nil {
			h.server.LoggerService.GetApplication().RecordCustomEvent(
				"healthCheckError", map[string]interface{}{
					"check_type":     "overall",
					"operation":      "health_check",
					"error_type":     "overall_unhealthy",
					"total_duration": time.Since(start).Milliseconds(),
				})
		}
		return c.JSON(http.StatusServiceUnavailable, response)
	}
	logger.Info().
		Dur("total_duration", time.Since(start)).
		Msg("O health check foi realizado com Sucesso")

	err := c.JSON(http.StatusOK, response)
	if err != nil {
		logger.Error().Err(err).Msg("Falha ao enviar a resposta do health check")
		if h.server.LoggerService != nil && h.server.LoggerService.GetApplication() != nil {
			h.server.LoggerService.GetApplication().RecordCustomEvent(
				"HealthCheckError", map[string]interface{}{
					"check_type":    "response",
					"operation":     "health_check",
					"error_type":    "response_error",
					"error_message": err.Error(),
				})
		}
		return fmt.Errorf("falha ao escrever a resposta JSON: %w", err)
	}
	return nil
}

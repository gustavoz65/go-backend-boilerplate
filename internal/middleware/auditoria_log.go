package middleware

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/model"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/repository"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/utils"
)

type AuditMiddleware struct {
	logger    *zerolog.Logger
	auditRepo repository.AuditLogRepository
}

func NewAuditMiddleware(repo repository.AuditLogRepository, logger *zerolog.Logger) *AuditMiddleware {
	return &AuditMiddleware{
		logger:    logger,
		auditRepo: repo,
	}
}

// Handler retorna o middleware que captura e registra todas as ações do usuário
// IMPORTANTE: Este middleware deve ser registrado DEPOIS do AuthMiddleware para ter acesso ao user_id
func (m *AuditMiddleware) Handler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Ignora rotas que não devem ser auditadas
			if m.shouldSkipAudit(c.Request().URL.Path) {
				return next(c)
			}

			// Captura informações do request ANTES de processar
			startTime := time.Now()
			r := c.Request()

			// Extrai informações do contexto (injetadas pelo AuthMiddleware)
			var userID *uuid.UUID
			if uid, ok := c.Get("user_id").(uuid.UUID); ok {
				userID = &uid
			}

			// Extrai informações do request
			ip := utils.GetRealIP(r)
			userAgent := utils.GetUserAgent(r)
			action := utils.GetActionFromRequest(r)
			entityType := utils.ExtractEntityFromPath(r.URL.Path)

			// Executa o próximo handler
			err := next(c)

			// Registra a auditoria de forma assíncrona para não bloquear a resposta
			go m.logAudit(userID, action, entityType, &ip, &userAgent)

			// Log de tempo de processamento (opcional)
			duration := time.Since(startTime)
			m.logger.Debug().
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Dur("duration", duration).
				Str("ip", ip).
				Msg("request processed")

			return err
		}
	}
}

// logAudit salva o log de auditoria no banco de dados
func (m *AuditMiddleware) logAudit(userID *uuid.UUID, action, entityType string, ip, userAgent *string) {
	ctx := context.Background()

	auditLog := &model.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		IPAddress:  ip,
		UserAgent:  userAgent,
		// OldValues e NewValues devem ser preenchidos pelos Services quando relevante
		// Este middleware apenas captura as informações básicas do request
	}

	if err := m.auditRepo.Create(ctx, auditLog); err != nil {
		m.logger.Error().
			Err(err).
			Str("action", action).
			Str("entity_type", entityType).
			Msg("failed to create audit log")
	}
}

// shouldSkipAudit determina se uma rota deve ser ignorada na auditoria
func (m *AuditMiddleware) shouldSkipAudit(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/api/v1/health",
		"/api/v1/auth/login",    // Login já é auditado internamente
		"/api/v1/auth/register", // Register já é auditado internamente
	}

	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}

	return false
}

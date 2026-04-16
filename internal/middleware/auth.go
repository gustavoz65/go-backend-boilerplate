// Pacote middleware fornece middlewares HTTP para a aplicação.
package middleware

import (
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/errs"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/model"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/service"
)

const (
	userIDKey = "user_id"
	emailKey  = "email"
	roleKey   = "role"
)

// AuthMiddleware valida o token JWT e injeta dados do usuario no contexto
func AuthMiddleware(authService *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return errs.NewUnauthorizedError("Token de autenticacao nao fornecido", false)
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return errs.NewUnauthorizedError("Formato de token invalido", false)
			}

			claims, err := authService.ValidateAccessToken(parts[1])
			if err != nil {
				if err == service.ErrTokenExpired {
					return errs.NewUnauthorizedError("Token expirado", false)
				}
				return errs.NewUnauthorizedError("Token invalido", false)
			}

			c.Set(userIDKey, claims.UserID)
			c.Set(emailKey, claims.Email)
			c.Set(roleKey, claims.Role)

			return next(c)
		}
	}
}

// WebSocketAuthMiddleware valida o token JWT via query parameter para WebSocket
func WebSocketAuthMiddleware(authService *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.QueryParam("token")
			if token == "" {
				return errs.NewUnauthorizedError("Token de autenticacao nao fornecido", false)
			}

			claims, err := authService.ValidateAccessToken(token)
			if err != nil {
				if err == service.ErrTokenExpired {
					return errs.NewUnauthorizedError("Token expirado", false)
				}
				return errs.NewUnauthorizedError("Token invalido", false)
			}

			c.Set(userIDKey, claims.UserID)
			c.Set(emailKey, claims.Email)
			c.Set(roleKey, claims.Role)

			return next(c)
		}
	}
}

// RequireRole verifica se o usuario tem uma das roles permitidas
func RequireRole(roles ...model.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get(roleKey).(model.UserRole)
			if !ok {
				return errs.NewUnauthorizedError("Nao autenticado", false)
			}

			for _, r := range roles {
				if role == r {
					return next(c)
				}
			}

			return errs.NewForbiddenError("Voce nao tem permissao para acessar este recurso", false)
		}
	}
}

// GetUserID retorna o ID do usuario autenticado
func GetUserID(c echo.Context) uuid.UUID {
	if id, ok := c.Get(userIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// GetUserEmail retorna o email do usuario autenticado
func GetUserEmail(c echo.Context) string {
	if email, ok := c.Get(emailKey).(string); ok {
		return email
	}
	return ""
}

// GetUserRole retorna a role do usuario autenticado
func GetUserRole(c echo.Context) model.UserRole {
	if role, ok := c.Get(roleKey).(model.UserRole); ok {
		return role
	}
	return ""
}

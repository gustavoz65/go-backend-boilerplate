package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/errs"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/model"
)

const (
	userIDKey = "user_id"
	emailKey  = "email"
	roleKey   = "role"
)

// RequireAuth valida o token JWT e injeta dados do usuario no contexto
func RequireAuth(authService *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			_ = c.Error(errs.NewUnauthorizedError("Token de autenticacao nao fornecido", false))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			_ = c.Error(errs.NewUnauthorizedError("Formato de token invalido", false))
			c.Abort()
			return
		}

		claims, err := authService.ValidateAccessToken(parts[1])
		if err != nil {
			if err == ErrTokenExpired {
				_ = c.Error(errs.NewUnauthorizedError("Token expirado", false))
			} else {
				_ = c.Error(errs.NewUnauthorizedError("Token invalido", false))
			}
			c.Abort()
			return
		}

		c.Set(userIDKey, claims.UserID)
		c.Set(emailKey, claims.Email)
		c.Set(roleKey, claims.Role)

		c.Next()
	}
}

// RequireAuthWebSocket valida o token JWT via query parameter para WebSocket
func RequireAuthWebSocket(authService *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			_ = c.Error(errs.NewUnauthorizedError("Token de autenticacao nao fornecido", false))
			c.Abort()
			return
		}

		claims, err := authService.ValidateAccessToken(token)
		if err != nil {
			if err == ErrTokenExpired {
				_ = c.Error(errs.NewUnauthorizedError("Token expirado", false))
			} else {
				_ = c.Error(errs.NewUnauthorizedError("Token invalido", false))
			}
			c.Abort()
			return
		}

		c.Set(userIDKey, claims.UserID)
		c.Set(emailKey, claims.Email)
		c.Set(roleKey, claims.Role)

		c.Next()
	}
}

// RequireRole verifica se o usuario tem uma das roles permitidas
func RequireRole(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, ok := c.Get(roleKey)
		role, isRole := roleVal.(model.UserRole)
		if !ok || !isRole {
			_ = c.Error(errs.NewUnauthorizedError("Nao autenticado", false))
			c.Abort()
			return
		}

		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		_ = c.Error(errs.NewForbiddenError("Voce nao tem permissao para acessar este recurso", false))
		c.Abort()
	}
}

// GetUserID retorna o ID do usuario autenticado
func GetUserID(c *gin.Context) uuid.UUID {
	if val, ok := c.Get(userIDKey); ok {
		if id, ok := val.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

// GetUserEmail retorna o email do usuario autenticado
func GetUserEmail(c *gin.Context) string {
	if val, ok := c.Get(emailKey); ok {
		if email, ok := val.(string); ok {
			return email
		}
	}
	return ""
}

// GetUserRole retorna a role do usuario autenticado
func GetUserRole(c *gin.Context) model.UserRole {
	if val, ok := c.Get(roleKey); ok {
		if role, ok := val.(model.UserRole); ok {
			return role
		}
	}
	return ""
}

package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/errs"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/modules/auth"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/repository"
)

// ErrorHandler converte o ultimo erro registrado via c.Error(...) em uma
// resposta HTTP estruturada. Deve ser registrado depois do LoggerMiddleware
// e antes de qualquer rota, para que o status final ja esteja escrito quando
// o LoggerMiddleware ler c.Writer.Status().
func ErrorHandler(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Written() {
			return
		}

		err := c.Errors.Last()
		if err == nil {
			return
		}

		writeError(c, logger, err.Err)
	}
}

func writeError(c *gin.Context, logger *zerolog.Logger, err error) {
	var httpErr *errs.HTTPError
	if errors.As(err, &httpErr) {
		c.JSON(httpErr.Status, httpErr)
		return
	}

	status, message := mapServiceError(err)
	if status != 0 {
		c.JSON(status, &errs.HTTPError{
			Code:    errs.MakeUpperCaseWithUnderscores(http.StatusText(status)),
			Message: message,
			Status:  status,
		})
		return
	}

	logger.Error().Err(err).
		Str("method", c.Request.Method).
		Str("path", c.Request.URL.Path).
		Msg("unhandled error")

	c.JSON(http.StatusInternalServerError, errs.NewInternalServerError())
}

func mapServiceError(err error) (int, string) {
	switch {
	// Erros de autenticacao
	case errors.Is(err, auth.ErrInvalidCredentials):
		return http.StatusUnauthorized, "Email ou senha invalidos"
	case errors.Is(err, auth.ErrUserNotActive):
		return http.StatusForbidden, "Conta de usuario inativa"
	case errors.Is(err, auth.ErrInvalidToken):
		return http.StatusUnauthorized, "Token invalido"
	case errors.Is(err, auth.ErrTokenExpired):
		return http.StatusUnauthorized, "Token expirado"
	case errors.Is(err, auth.ErrPasswordMismatch):
		return http.StatusBadRequest, "Senha atual incorreta"

	// Erros de repositorio - Not Found
	case errors.Is(err, repository.ErrUserNotFound):
		return http.StatusNotFound, "Usuario nao encontrado"
	case errors.Is(err, repository.ErrNotificationNotFound):
		return http.StatusNotFound, "Notificacao nao encontrada"

	// Erros de repositorio - Conflict
	case errors.Is(err, repository.ErrUserAlreadyExists):
		return http.StatusConflict, "Email ja cadastrado"

	// Erros de Firebase / social login
	case errors.Is(err, auth.ErrFirebaseNotConfigured):
		return http.StatusServiceUnavailable, "Login social indisponível no momento"
	case errors.Is(err, auth.ErrInvalidIDToken):
		return http.StatusUnauthorized, "Token do provedor inválido"
	case errors.Is(err, auth.ErrCannotUnlinkLastProvider):
		return http.StatusBadRequest, "Não é possível desvincular o último método de autenticação"

	// Erros de sessao
	case errors.Is(err, repository.ErrSessionNotFound),
		errors.Is(err, repository.ErrSessionExpired),
		errors.Is(err, repository.ErrSessionRevoked):
		return http.StatusUnauthorized, "Sessao invalida"

	default:
		return 0, ""
	}
}

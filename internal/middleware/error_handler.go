package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/errs"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/repository"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/service"
)

// ErrorHandler converte erros em respostas HTTP estruturadas
func ErrorHandler(logger *zerolog.Logger) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		// Se ja e um HTTPError do nosso pacote
		var httpErr *errs.HTTPError
		if errors.As(err, &httpErr) {
			_ = c.JSON(httpErr.Status, httpErr)
			return
		}

		// Erros do Echo
		var echoErr *echo.HTTPError
		if errors.As(err, &echoErr) {
			msg, _ := echoErr.Message.(string)
			if msg == "" {
				msg = http.StatusText(echoErr.Code)
			}
			_ = c.JSON(echoErr.Code, &errs.HTTPError{
				Code:    errs.MakeUpperCaseWithUnderscores(http.StatusText(echoErr.Code)),
				Message: msg,
				Status:  echoErr.Code,
			})
			return
		}

		// Mapear erros de servico/repositorio para HTTP
		status, message := mapServiceError(err)
		if status != 0 {
			_ = c.JSON(status, &errs.HTTPError{
				Code:    errs.MakeUpperCaseWithUnderscores(http.StatusText(status)),
				Message: message,
				Status:  status,
			})
			return
		}

		// Erro interno nao mapeado
		logger.Error().Err(err).
			Str("method", c.Request().Method).
			Str("path", c.Request().URL.Path).
			Msg("unhandled error")

		_ = c.JSON(http.StatusInternalServerError, errs.NewInternalServerError())
	}
}

func mapServiceError(err error) (int, string) {
	switch {
	// Erros de autenticacao
	case errors.Is(err, service.ErrInvalidCredentials):
		return http.StatusUnauthorized, "Email ou senha invalidos"
	case errors.Is(err, service.ErrUserNotActive):
		return http.StatusForbidden, "Conta de usuario inativa"
	case errors.Is(err, service.ErrInvalidToken):
		return http.StatusUnauthorized, "Token invalido"
	case errors.Is(err, service.ErrTokenExpired):
		return http.StatusUnauthorized, "Token expirado"
	case errors.Is(err, service.ErrPasswordMismatch):
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
	case errors.Is(err, service.ErrFirebaseNotConfigured):
		return http.StatusServiceUnavailable, "Login social indisponível no momento"
	case errors.Is(err, service.ErrInvalidIDToken):
		return http.StatusUnauthorized, "Token do provedor inválido"
	case errors.Is(err, service.ErrCannotUnlinkLastProvider):
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

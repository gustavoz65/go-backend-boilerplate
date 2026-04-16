package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/example/go-backend-boilerplate/internal/errs"
)

// RecoveryMiddleware captura panics e retorna erro 500
func RecoveryMiddleware(logger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					var stackBuf [4096]byte
					n := runtime.Stack(stackBuf[:], false)

					logger.Error().
						Str("method", c.Request().Method).
						Str("path", c.Request().URL.Path).
						Str("remote_ip", c.RealIP()).
						Str("stack", string(stackBuf[:n])).
						Msgf("PANIC RECOVERED: %v", r)

					httpErr := errs.NewInternalServerError()
					_ = c.JSON(http.StatusInternalServerError, httpErr)
				}
			}()

			return next(c)
		}
	}
}

// RecoveryMiddlewareWithMessage retorna uma mensagem customizada
func RecoveryMiddlewareWithMessage(logger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					logger.Error().
						Err(err).
						Str("method", c.Request().Method).
						Str("path", c.Request().URL.Path).
						Msg("panic recovered")

					httpErr := errs.NewInternalServerError()
					_ = c.JSON(http.StatusInternalServerError, httpErr)
				}
			}()

			return next(c)
		}
	}
}

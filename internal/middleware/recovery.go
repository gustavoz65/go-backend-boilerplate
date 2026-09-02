package middleware

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/errs"
)

// RecoveryMiddleware captura panics e retorna erro 500
func RecoveryMiddleware(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var stackBuf [4096]byte
				n := runtime.Stack(stackBuf[:], false)

				logger.Error().
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Str("remote_ip", c.ClientIP()).
					Str("stack", string(stackBuf[:n])).
					Msgf("PANIC RECOVERED: %v", r)

				c.AbortWithStatusJSON(http.StatusInternalServerError, errs.NewInternalServerError())
			}
		}()

		c.Next()
	}
}

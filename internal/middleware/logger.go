package middleware

import (
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

const loggerKey = "logger"

var sensitiveQueryParams = []string{
	"token",
	"password",
	"secret",
	"api_key",
	"apikey",
	"access_token",
	"refresh_token",
	"id_token",
}

func LoggerMiddleware(logger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()

			// Sanitiza query params para não logar dados sensíveis
			sanitizedQuery := sanitizeQueryParams(req.URL.RawQuery)

			// Injeta logger no contexto
			requestLogger := logger.With().
				Str("request_id", c.Response().Header().Get(echo.HeaderXRequestID)).
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Str("query", sanitizedQuery).
				Str("remote_ip", c.RealIP()).
				Str("user_agent", sanitizeUserAgent(req.UserAgent())).
				Logger()

			c.Set(loggerKey, &requestLogger)

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status

			logEvent := requestLogger.Info()
			if status >= 500 {
				logEvent = requestLogger.Error()
			} else if status >= 400 {
				logEvent = requestLogger.Warn()
			}

			logEvent.
				Int("status", status).
				Dur("duration", duration).
				Int64("bytes_out", c.Response().Size).
				Msg("request completed")

			return err
		}
	}
}

func sanitizeQueryParams(query string) string {
	if query == "" {
		return ""
	}

	var sanitized []string
	params := strings.Split(query, "&")

	for _, param := range params {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			sanitized = append(sanitized, param)
			continue
		}

		key := strings.ToLower(parts[0])
		isSensitive := false

		for _, sensitive := range sensitiveQueryParams {
			if strings.Contains(key, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized = append(sanitized, parts[0]+"=[REDACTED]")
		} else {
			sanitized = append(sanitized, param)
		}
	}

	return strings.Join(sanitized, "&")
}

func sanitizeUserAgent(ua string) string {
	const maxLength = 200
	if len(ua) > maxLength {
		return ua[:maxLength] + "..."
	}
	return ua
}

// GetLogger retorna o logger do contexto
func GetLogger(c echo.Context) *zerolog.Logger {
	if logger, ok := c.Get(loggerKey).(*zerolog.Logger); ok {
		return logger
	}
	l := zerolog.Nop()
	return &l
}

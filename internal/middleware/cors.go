package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// CORSMiddleware configura CORS baseado nas origens permitidas
func CORSMiddleware(allowedOrigins []string) echo.MiddlewareFunc {
	var parsedOrigins []string
	for _, origin := range allowedOrigins {
		if strings.Contains(origin, ",") || strings.Contains(origin, " ") {
			parts := strings.FieldsFunc(origin, func(r rune) bool {
				return r == ',' || r == ' '
			})
			parsedOrigins = append(parsedOrigins, parts...)
		} else if strings.TrimSpace(origin) != "" {
			parsedOrigins = append(parsedOrigins, strings.TrimSpace(origin))
		}
	}

	if len(parsedOrigins) == 0 {
		parsedOrigins = []string{"http://localhost:3000", "http://localhost:4000"}
	}

	allowedOrigins = parsedOrigins
	allowCredentials := true
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowCredentials = false
			break
		}
	}

	return echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Requested-With",
			"X-CSRF-Token",
			// Headers necessários para WebSocket
			"Sec-WebSocket-Key",
			"Sec-WebSocket-Version",
			"Sec-WebSocket-Protocol",
			"Sec-WebSocket-Extensions",
			"Connection",
			"Upgrade",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-CSRF-Token",
		},
		AllowCredentials: allowCredentials,
		MaxAge:           86400,
	})
}

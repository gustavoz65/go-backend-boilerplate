package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var corsAllowedMethods = strings.Join([]string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodOptions,
}, ", ")

var corsAllowedHeaders = strings.Join([]string{
	"Origin",
	"Content-Type",
	"Accept",
	"Authorization",
	"X-Requested-With",
	"X-CSRF-Token",
	// Headers necessários para WebSocket
	"Sec-WebSocket-Key",
	"Sec-WebSocket-Version",
	"Sec-WebSocket-Protocol",
	"Sec-WebSocket-Extensions",
	"Connection",
	"Upgrade",
}, ", ")

var corsExposeHeaders = strings.Join([]string{
	"Content-Length",
	"Content-Type",
	"X-CSRF-Token",
}, ", ")

const corsMaxAgeSeconds = 86400

// CORSMiddleware configura CORS baseado nas origens permitidas
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
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

	allowAll := false
	originSet := make(map[string]struct{}, len(parsedOrigins))
	for _, origin := range parsedOrigins {
		if origin == "*" {
			allowAll = true
		}
		originSet[origin] = struct{}{}
	}
	allowCredentials := !allowAll

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		allowOrigin := ""
		if allowAll {
			allowOrigin = "*"
		} else if origin != "" {
			if _, ok := originSet[origin]; ok {
				allowOrigin = origin
			}
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			c.Header("Vary", "Origin")
			if allowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			c.Header("Access-Control-Expose-Headers", corsExposeHeaders)
		}

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", corsAllowedMethods)
			c.Header("Access-Control-Allow-Headers", corsAllowedHeaders)
			c.Header("Access-Control-Max-Age", strconv.Itoa(corsMaxAgeSeconds))
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/errs"
)

const (
	csrfTokenHeader  = "X-CSRF-Token"
	csrfCookieName   = "csrf_token"
	csrfTokenLength  = 32
	csrfCookieMaxAge = 86400 // 24 horas
)

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		cookieToken, _ := c.Cookie(csrfCookieName)

		if cookieToken == "" {
			cookieToken = generateCSRFToken()
			setCSRFCookie(c, cookieToken)
		}

		headerToken := c.GetHeader(csrfTokenHeader)

		if headerToken == "" || !secureCompare(cookieToken, headerToken) {
			_ = c.Error(errs.NewForbiddenError("Token CSRF invalido ou ausente", false))
			c.Abort()
			return
		}

		c.Next()
	}
}

func CSRFTokenGenerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(csrfCookieName)
		if err != nil || token == "" {
			token = generateCSRFToken()
			setCSRFCookie(c, token)
		}

		// serve o token no header para que o cliente possa ler e usar nas requisições subsequentes
		c.Header(csrfTokenHeader, token)

		c.Next()
	}
}

func generateCSRFToken() string {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		return base64.URLEncoding.EncodeToString([]byte(time.Now().String()))
	}
	return base64.URLEncoding.EncodeToString(b)
}

func setCSRFCookie(c *gin.Context, token string) {
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteStrictMode
	if isSecure {
		sameSite = http.SameSiteNoneMode
	}
	c.SetSameSite(sameSite)
	c.SetCookie(csrfCookieName, token, csrfCookieMaxAge, "/", "", isSecure, false)
}

func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

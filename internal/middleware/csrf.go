package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/errs"
)

const (
	csrfTokenHeader  = "X-CSRF-Token"
	csrfCookieName   = "csrf_token"
	csrfTokenLength  = 32
	csrfCookieMaxAge = 86400 // 24 horas
)

func CSRFMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodGet ||
				c.Request().Method == http.MethodHead ||
				c.Request().Method == http.MethodOptions {
				return next(c)
			}

			cookie, err := c.Cookie(csrfCookieName)
			var cookieToken string
			if err == nil {
				cookieToken = cookie.Value
			}

			if cookieToken == "" {
				cookieToken = generateCSRFToken()
				setCSRFCookie(c, cookieToken)
			}

			headerToken := c.Request().Header.Get(csrfTokenHeader)

			if headerToken == "" || !secureCompare(cookieToken, headerToken) {
				return errs.NewForbiddenError("Token CSRF invalido ou ausente", false)
			}

			return next(c)
		}
	}
}

func CSRFTokenGenerator() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(csrfCookieName)
			var token string
			if err != nil || cookie.Value == "" {
				token = generateCSRFToken()
				setCSRFCookie(c, token)
			} else {
				token = cookie.Value
			}

			// serve o token no header para que o cliente possa ler e usar nas requisições subsequentes
			c.Response().Header().Set(csrfTokenHeader, token)

			return next(c)
		}
	}
}

func generateCSRFToken() string {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		return base64.URLEncoding.EncodeToString([]byte(time.Now().String()))
	}
	return base64.URLEncoding.EncodeToString(b)
}

func setCSRFCookie(c echo.Context, token string) {
	isSecure := c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteStrictMode
	if isSecure {
		sameSite = http.SameSiteNoneMode
	}
	cookie := &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   csrfCookieMaxAge,
		HttpOnly: false,
		Secure:   isSecure,
		SameSite: sameSite,
	}
	c.SetCookie(cookie)
}

func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

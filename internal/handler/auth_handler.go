package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/middleware"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/model"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/service"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/validation"
)

const (
	refreshTokenCookieName   = "refresh_token"
	refreshTokenCookieMaxAge = 7 * 24 * 60 * 60 // 7 dias em segundos
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func setRefreshTokenCookie(c echo.Context, token string) {
	cookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   refreshTokenCookieMaxAge,
		HttpOnly: true,
		Secure:   c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}

func clearRefreshTokenCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}

func getRefreshTokenFromRequest(c echo.Context) string {
	cookie, err := c.Cookie(refreshTokenCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	var req model.RefreshTokenRequest
	if err := c.Bind(&req); err == nil && req.RefreshToken != "" {
		return req.RefreshToken
	}

	return ""
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req model.RegisterRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		return err
	}

	response, err := h.authService.Register(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	setRefreshTokenCookie(c, response.RefreshToken)

	return c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		return err
	}

	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	response, err := h.authService.Login(c.Request().Context(), &req, ipAddress, userAgent)
	if err != nil {
		return err
	}

	setRefreshTokenCookie(c, response.RefreshToken)

	return c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	refreshToken := getRefreshTokenFromRequest(c)
	if refreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh token não fornecido")
	}

	response, err := h.authService.RefreshToken(c.Request().Context(), refreshToken)
	if err != nil {
		return err
	}

	setRefreshTokenCookie(c, response.RefreshToken)

	return c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	refreshToken := getRefreshTokenFromRequest(c)
	if refreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh token não fornecido")
	}

	if err := h.authService.Logout(c.Request().Context(), refreshToken); err != nil {
		return err
	}

	clearRefreshTokenCookie(c)

	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) ChangePassword(c echo.Context) error {
	var req model.ChangePasswordRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		return err
	}

	userID := middleware.GetUserID(c)
	if err := h.authService.ChangePassword(c.Request().Context(), userID, &req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) SetPassword(c echo.Context) error {
	var req model.SetPasswordRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		return err
	}

	userID := middleware.GetUserID(c)
	if err := h.authService.SetPassword(c.Request().Context(), userID, &req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// Social Login Handlers

func (h *AuthHandler) SocialLogin(c echo.Context) error {
	var req model.SocialLoginRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		return err
	}

	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	response, err := h.authService.SocialLogin(c.Request().Context(), &req, ipAddress, userAgent)
	if err != nil {
		return err
	}

	setRefreshTokenCookie(c, response.RefreshToken)

	return c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) LinkProvider(c echo.Context) error {
	var req model.LinkProviderRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		return err
	}

	userID := middleware.GetUserID(c)
	if err := h.authService.LinkProvider(c.Request().Context(), userID, &req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) UnlinkProvider(c echo.Context) error {
	provider := c.Param("provider")
	if provider == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "provider parameter is required")
	}

	userID := middleware.GetUserID(c)
	if err := h.authService.UnlinkProvider(c.Request().Context(), userID, provider); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) GetLinkedProviders(c echo.Context) error {
	userID := middleware.GetUserID(c)

	response, err := h.authService.GetLinkedProviders(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response)
}

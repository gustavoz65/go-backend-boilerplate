package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/errs"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/model"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/validation"
)

const (
	refreshTokenCookieName   = "refresh_token"
	refreshTokenCookieMaxAge = 7 * 24 * 60 * 60 // 7 dias em segundos
)

type Handler struct {
	authService *Service
}

func NewHandler(authService *Service) *Handler {
	return &Handler{authService: authService}
}

func setRefreshTokenCookie(c *gin.Context, token string) {
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshTokenCookieName, token, refreshTokenCookieMaxAge, "/", "", isSecure, true)
}

func clearRefreshTokenCookie(c *gin.Context) {
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshTokenCookieName, "", -1, "/", "", isSecure, true)
}

func getRefreshTokenFromRequest(c *gin.Context) string {
	if cookie, err := c.Cookie(refreshTokenCookieName); err == nil && cookie != "" {
		return cookie
	}

	var req model.RefreshTokenRequest
	if err := c.ShouldBind(&req); err == nil && req.RefreshToken != "" {
		return req.RefreshToken
	}

	return ""
}

func (h *Handler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	response, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	setRefreshTokenCookie(c, response.RefreshToken)

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	response, err := h.authService.Login(c.Request.Context(), &req, ipAddress, userAgent)
	if err != nil {
		_ = c.Error(err)
		return
	}

	setRefreshTokenCookie(c, response.RefreshToken)

	c.JSON(http.StatusOK, response)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	refreshToken := getRefreshTokenFromRequest(c)
	if refreshToken == "" {
		_ = c.Error(errs.NewBadRequestError("refresh token não fornecido", false, nil, nil, nil))
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		_ = c.Error(err)
		return
	}

	setRefreshTokenCookie(c, response.RefreshToken)

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Logout(c *gin.Context) {
	refreshToken := getRefreshTokenFromRequest(c)
	if refreshToken == "" {
		_ = c.Error(errs.NewBadRequestError("refresh token não fornecido", false, nil, nil, nil))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), refreshToken); err != nil {
		_ = c.Error(err)
		return
	}

	clearRefreshTokenCookie(c)

	c.Status(http.StatusNoContent)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	var req model.ChangePasswordRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	userID := GetUserID(c)
	if err := h.authService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) SetPassword(c *gin.Context) {
	var req model.SetPasswordRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	userID := GetUserID(c)
	if err := h.authService.SetPassword(c.Request.Context(), userID, &req); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Social Login Handlers

func (h *Handler) SocialLogin(c *gin.Context) {
	var req model.SocialLoginRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	response, err := h.authService.SocialLogin(c.Request.Context(), &req, ipAddress, userAgent)
	if err != nil {
		_ = c.Error(err)
		return
	}

	setRefreshTokenCookie(c, response.RefreshToken)

	c.JSON(http.StatusOK, response)
}

func (h *Handler) LinkProvider(c *gin.Context) {
	var req model.LinkProviderRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	userID := GetUserID(c)
	if err := h.authService.LinkProvider(c.Request.Context(), userID, &req); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) UnlinkProvider(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		_ = c.Error(errs.NewBadRequestError("provider parameter is required", false, nil, nil, nil))
		return
	}

	userID := GetUserID(c)
	if err := h.authService.UnlinkProvider(c.Request.Context(), userID, provider); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetLinkedProviders(c *gin.Context) {
	userID := GetUserID(c)

	response, err := h.authService.GetLinkedProviders(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}

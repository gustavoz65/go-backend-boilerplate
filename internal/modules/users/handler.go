package users

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/model"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/modules/auth"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/validation"
)

type Handler struct {
	userService *Service
}

func NewHandler(userService *Service) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) GetMe(c *gin.Context) {
	userID := auth.GetUserID(c)

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) UpdateMe(c *gin.Context) {
	var req model.UpdateUserRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	userID := auth.GetUserID(c)

	user, err := h.userService.Update(c.Request.Context(), userID, &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) DeactivateMe(c *gin.Context) {
	userID := auth.GetUserID(c)

	if err := h.userService.Deactivate(c.Request.Context(), userID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetSettings(c *gin.Context) {
	userID := auth.GetUserID(c)

	settings, err := h.userService.GetSettings(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var req model.UpdateUserSettingsRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	userID := auth.GetUserID(c)

	settings, err := h.userService.UpdateSettings(c.Request.Context(), userID, &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) CompleteOnboarding(c *gin.Context) {
	userID := auth.GetUserID(c)

	user, err := h.userService.CompleteOnboarding(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

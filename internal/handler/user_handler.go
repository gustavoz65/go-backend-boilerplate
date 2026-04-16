package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/middleware"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/model"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/service"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/validation"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(c echo.Context) error {
	userID := middleware.GetUserID(c)

	user, err := h.userService.GetByID(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateMe(c echo.Context) error {
	var req model.UpdateUserRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		return err
	}

	userID := middleware.GetUserID(c)

	user, err := h.userService.Update(c.Request().Context(), userID, &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeactivateMe(c echo.Context) error {
	userID := middleware.GetUserID(c)

	if err := h.userService.Deactivate(c.Request().Context(), userID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *UserHandler) GetSettings(c echo.Context) error {
	userID := middleware.GetUserID(c)

	settings, err := h.userService.GetSettings(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, settings)
}

func (h *UserHandler) UpdateSettings(c echo.Context) error {
	var req model.UpdateUserSettingsRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		return err
	}

	userID := middleware.GetUserID(c)

	settings, err := h.userService.UpdateSettings(c.Request().Context(), userID, &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, settings)
}

func (h *UserHandler) CompleteOnboarding(c echo.Context) error {
	userID := middleware.GetUserID(c)

	user, err := h.userService.CompleteOnboarding(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}

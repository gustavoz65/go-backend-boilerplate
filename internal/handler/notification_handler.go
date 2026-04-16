package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/example/go-backend-boilerplate/internal/errs"
	"github.com/example/go-backend-boilerplate/internal/middleware"
	"github.com/example/go-backend-boilerplate/internal/service"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

func (h *NotificationHandler) GetAll(c echo.Context) error {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.notificationService.GetAll(c.Request().Context(), userID, page, pageSize)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, result)
}

func (h *NotificationHandler) GetUnread(c echo.Context) error {
	userID := middleware.GetUserID(c)

	notifications, err := h.notificationService.GetUnread(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) GetUnreadCount(c echo.Context) error {
	userID := middleware.GetUserID(c)

	count, err := h.notificationService.GetUnreadCount(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]int64{
		"count": count,
	})
}

func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	userID := middleware.GetUserID(c)

	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.NewBadRequestError("ID de notificacao invalido", false, nil, nil, nil)
	}

	if err := h.notificationService.MarkAsRead(c.Request().Context(), userID, notificationID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {
	userID := middleware.GetUserID(c)

	if err := h.notificationService.MarkAllAsRead(c.Request().Context(), userID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *NotificationHandler) Delete(c echo.Context) error {
	userID := middleware.GetUserID(c)

	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errs.NewBadRequestError("ID de notificacao invalido", false, nil, nil, nil)
	}

	if err := h.notificationService.Delete(c.Request().Context(), userID, notificationID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

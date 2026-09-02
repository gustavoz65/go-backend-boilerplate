package notifications

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/errs"
	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/modules/auth"
)

type Handler struct {
	notificationService *Service
}

func NewHandler(notificationService *Service) *Handler {
	return &Handler{notificationService: notificationService}
}

func (h *Handler) GetAll(c *gin.Context) {
	userID := auth.GetUserID(c)

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.notificationService.GetAll(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetUnread(c *gin.Context) {
	userID := auth.GetUserID(c)

	notifications, err := h.notificationService.GetUnread(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID := auth.GetUserID(c)

	count, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, map[string]int64{
		"count": count,
	})
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	userID := auth.GetUserID(c)

	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errs.NewBadRequestError("ID de notificacao invalido", false, nil, nil, nil))
		return
	}

	if err := h.notificationService.MarkAsRead(c.Request.Context(), userID, notificationID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID := auth.GetUserID(c)

	if err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := auth.GetUserID(c)

	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(errs.NewBadRequestError("ID de notificacao invalido", false, nil, nil, nil))
		return
	}

	if err := h.notificationService.Delete(c.Request.Context(), userID, notificationID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

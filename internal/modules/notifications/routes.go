package notifications

import "github.com/gin-gonic/gin"

// Middlewares agrupa os middlewares compartilhados que as rotas de notifications precisam.
type Middlewares struct {
	RequireAuth gin.HandlerFunc
	Audit       gin.HandlerFunc
	CSRF        gin.HandlerFunc
	MutationRL  gin.HandlerFunc
	ReadRL      gin.HandlerFunc
}

// RegisterRoutes registra todas as rotas do módulo de notificações em /notifications.
func RegisterRoutes(api *gin.RouterGroup, h *Handler, mw Middlewares) {
	notifications := api.Group("/notifications", mw.RequireAuth, mw.Audit, mw.CSRF)

	notifications.GET("", mw.ReadRL, h.GetAll)
	notifications.GET("/unread", mw.ReadRL, h.GetUnread)
	notifications.GET("/unread/count", mw.ReadRL, h.GetUnreadCount)
	notifications.PATCH("/:id/read", mw.MutationRL, h.MarkAsRead)
	notifications.PATCH("/read-all", mw.MutationRL, h.MarkAllAsRead)
	notifications.DELETE("/:id", mw.MutationRL, h.Delete)
}

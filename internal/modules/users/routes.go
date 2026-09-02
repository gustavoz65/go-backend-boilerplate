package users

import "github.com/gin-gonic/gin"

// Middlewares agrupa os middlewares compartilhados que as rotas de users precisam.
type Middlewares struct {
	RequireAuth gin.HandlerFunc
	Audit       gin.HandlerFunc
	CSRF        gin.HandlerFunc
	MutationRL  gin.HandlerFunc
	ReadRL      gin.HandlerFunc
}

// RegisterRoutes registra todas as rotas do módulo de usuários em /users.
func RegisterRoutes(api *gin.RouterGroup, h *Handler, mw Middlewares) {
	users := api.Group("/users", mw.RequireAuth, mw.Audit, mw.CSRF)

	users.GET("/me", mw.ReadRL, h.GetMe)
	users.PUT("/me", mw.MutationRL, h.UpdateMe)
	users.DELETE("/me", mw.MutationRL, h.DeactivateMe)
	users.GET("/settings", mw.ReadRL, h.GetSettings)
	users.PUT("/settings", mw.MutationRL, h.UpdateSettings)
	users.PATCH("/me/onboarding-complete", mw.MutationRL, h.CompleteOnboarding)
}

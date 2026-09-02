package auth

import "github.com/gin-gonic/gin"

// Middlewares agrupa os middlewares compartilhados que as rotas de auth precisam.
// Cada instância é construída uma única vez na composição do router (internal/router)
// e repassada aqui, para que este módulo não precise conhecer Redis, JWT, etc.
type Middlewares struct {
	AuthRateLimit    gin.HandlerFunc
	RefreshRateLimit gin.HandlerFunc
	CSRFTokenGen     gin.HandlerFunc
	RequireAuth      gin.HandlerFunc
	Audit            gin.HandlerFunc
	CSRF             gin.HandlerFunc
	MutationRL       gin.HandlerFunc
	ReadRL           gin.HandlerFunc
}

// RegisterRoutes registra todas as rotas do módulo de autenticação em /auth.
func RegisterRoutes(api *gin.RouterGroup, h *Handler, mw Middlewares) {
	public := api.Group("/auth", mw.AuthRateLimit, mw.CSRFTokenGen)
	public.POST("/register", h.Register)
	public.POST("/login", h.Login)
	public.POST("/social/login", h.SocialLogin)

	refresh := api.Group("/auth", mw.RefreshRateLimit, mw.CSRFTokenGen)
	refresh.POST("/refresh", h.RefreshToken)
	refresh.POST("/logout", h.Logout)

	protected := api.Group("/auth", mw.RequireAuth, mw.Audit, mw.CSRF)
	protected.POST("/change-password", mw.MutationRL, h.ChangePassword)
	protected.POST("/set-password", mw.MutationRL, h.SetPassword)
	protected.POST("/social/link", mw.MutationRL, h.LinkProvider)
	protected.DELETE("/social/:provider", mw.MutationRL, h.UnlinkProvider)
	protected.GET("/social/providers", mw.ReadRL, h.GetLinkedProviders)
}

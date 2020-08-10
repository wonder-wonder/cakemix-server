package handler

import "github.com/gin-gonic/gin"

// AuthHandler is handlers of auth
func (h *Handler) AuthHandler(r *gin.RouterGroup) {
	auth := r.Group("auth")
	auth.POST("login", h.notimplHandler)
	auth.POST("regist/pre/:token", h.notimplHandler)
	auth.POST("regist/verify/:token", h.notimplHandler)
	auth.POST("pass/reset", h.notimplHandler)
	auth.GET("pass/reset/verify/:token", h.notimplHandler)
	auth.POST("pass/reset/verify/:token", h.notimplHandler)

	authck := auth.Group("/", h.CheckAuthMiddleware())
	authck.POST("logout", h.notimplHandler)
	authck.GET("check/token", h.notimplHandler)
	authck.GET("regist/gen/token", h.notimplHandler)
	authck.POST("pass/change", h.notimplHandler)
}

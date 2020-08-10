package handler

import "github.com/gin-gonic/gin"

// ProfileHandler is handlers of profile
func (h *Handler) ProfileHandler(r *gin.RouterGroup) {
	profck := r.Group("profile", h.CheckAuthMiddleware())
	profck.GET(":name", h.notimplHandler)
	profck.PUT(":name", h.notimplHandler)
}

package handler

import "github.com/gin-gonic/gin"

// TeamHandler is handlers of team
func (h *Handler) TeamHandler(r *gin.RouterGroup) {
	teamck := r.Group("team", h.CheckAuthMiddleware())
	teamck.GET(":teamid/member", h.notimplHandler)
	teamck.POST(":teamid/member", h.notimplHandler)
	teamck.PUT(":teamid/member", h.notimplHandler)
	teamck.DELETE(":teamid/member", h.notimplHandler)
	teamck.POST("/", h.notimplHandler)
	teamck.DELETE(":teamid", h.notimplHandler)
}

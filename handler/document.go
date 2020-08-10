package handler

import "github.com/gin-gonic/gin"

// DocumentHandler is handlers of documents
func (h *Handler) DocumentHandler(r *gin.RouterGroup) {
	docck := r.Group("doc", h.CheckAuthMiddleware())
	docck.POST(":folderid", h.notimplHandler)
	docck.DELETE(":docid", h.notimplHandler)
	docck.PUT(":docid/move/:folderid", h.notimplHandler)
}

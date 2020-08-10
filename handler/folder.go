package handler

import "github.com/gin-gonic/gin"

// FolderHandler is handlers of folders
func (h *Handler) FolderHandler(r *gin.RouterGroup) {
	folderck := r.Group("folder", h.CheckAuthMiddleware())
	folderck.GET(":folderid", h.notimplHandler)
	folderck.POST(":folderid", h.notimplHandler)
	folderck.DELETE(":folderid", h.notimplHandler)
	folderck.PUT(":folderid/move/:targetfid", h.notimplHandler)
}

package handler

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/model"
)

// ImageHandler is handlers of profile
func (h *Handler) ImageHandler(r *gin.RouterGroup) {
	imgck := r.Group("image", h.CheckAuthMiddleware())
	imgck.POST("", h.uploadImageHandler)
	img := r.Group("image")
	img.GET(":id", h.getImageHandler)
}

func (h *Handler) getImageHandler(c *gin.Context) {
	imgid := c.Param("id")

	c.File(path.Join(dataDir, imageDir, imgid))
}

func (h *Handler) uploadImageHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	imgid, err := db.GenerateID(db.IDTypeImageID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Upload the file to specific dst.
	err = c.SaveUploadedFile(file, path.Join(dataDir, imageDir, imgid))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, model.ImageUploadRes{ID: imgid})
}

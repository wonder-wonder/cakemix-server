package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/model"
)

// DocumentHandler is handlers of documents
func (h *Handler) DocumentHandler(r *gin.RouterGroup) {
	r.GET("doc/:docid/ws", h.getOTHandler)
	docck := r.Group("doc", h.CheckAuthMiddleware())
	docck.POST(":folderid", h.createDocumentHandler)
	docck.DELETE(":docid", h.deleteDocumentHandler)
	docck.PUT(":docid/move/:folderid", h.moveDocumentHandler)
}
func (h *Handler) createDocumentHandler(c *gin.Context) {
	parentfid := c.Param("folderid")

	finfo, err := h.db.GetFolderInfo(parentfid)
	if err != nil {
		if err == db.ErrFolderNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if !isRelatedUUID(c, finfo.OwnerUUID) && finfo.Permission != db.FilePermReadWrite {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	did, err := h.db.CreateDocument("Untitled", db.FilePermPrivate, parentfid, finfo.OwnerUUID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatusJSON(http.StatusOK, model.CreateDocumentRes{DocumentID: did})
}

func (h *Handler) deleteDocumentHandler(c *gin.Context) {
	did := c.Param("docid")

	dinfo, err := h.db.GetDocumentInfo(did)
	if err != nil {
		if err == db.ErrFolderNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !isRelatedUUID(c, dinfo.OwnerUUID) && dinfo.Permission != db.FilePermReadWrite {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = h.db.DeleteDocument(did)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}
func (h *Handler) moveDocumentHandler(c *gin.Context) {
	did := c.Param("docid")
	targetfid := c.Param("targetfid")

	// Check document permission
	dinfo, err := h.db.GetFolderInfo(did)
	if err != nil {
		if err == db.ErrFolderNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !isRelatedUUID(c, dinfo.OwnerUUID) && dinfo.Permission != db.FilePermReadWrite {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Check original parent folder permission
	finfo, err := h.db.GetFolderInfo(dinfo.ParentFolderUUID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !isRelatedUUID(c, finfo.OwnerUUID) && finfo.Permission != db.FilePermReadWrite {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Check target folder permission
	finfo, err = h.db.GetFolderInfo(targetfid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !isRelatedUUID(c, finfo.OwnerUUID) && finfo.Permission != db.FilePermReadWrite {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = h.db.MoveDocument(did, targetfid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

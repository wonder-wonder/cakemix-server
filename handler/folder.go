package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/model"
)

// FolderHandler is handlers of folders
func (h *Handler) FolderHandler(r *gin.RouterGroup) {
	folderck := r.Group("folder", h.CheckAuthMiddleware())
	folderck.GET(":folderid", h.getFolderHandler)
	folderck.POST(":folderid", h.createFolderHandler)
	folderck.DELETE(":folderid", h.deleteFolderHandler)
	folderck.PUT(":folderid/move/:targetfid", h.moveFolderHandler)
}
func (h *Handler) getFolderHandler(c *gin.Context) {
	fid := c.Param("folderid")
	listtype := c.Query("type")

	follist := []model.Folder{}
	doclist := []model.Document{}

	//TODO: permission check

	if listtype == "" || listtype == "folder" {
		folidlist, err := h.db.GetFolderList(fid)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		for _, v := range folidlist {
			folinfo, err := h.db.GetFolderInfo(v)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			ownp, err := h.db.GetProfile(folinfo.OwnerUUID)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			updp, err := h.db.GetProfile(folinfo.UpdaterUUID)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			follist = append(follist, model.Folder{
				Owner: model.Profile{
					UUID:    ownp.UUID,
					Name:    ownp.Name,
					IconURI: ownp.IconURI,
					Attr:    ownp.Attr,
					IsTeam:  (ownp.UUID[0] == 't'),
				},
				Updater: model.Profile{
					UUID:    updp.UUID,
					Name:    updp.Name,
					IconURI: updp.IconURI,
					Attr:    updp.Attr,
					IsTeam:  (updp.UUID[0] == 't'),
				},
				Name:       folinfo.Name,
				Permission: folinfo.Permission,
				CreatedAt:  folinfo.CreatedAt,
				UpdatedAt:  folinfo.UpdatedAt,
			})
		}
	}
	if listtype == "" || listtype == "document" {
		docidlist, err := h.db.GetDocList(fid)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		for _, v := range docidlist {
			docinfo, err := h.db.GetDocumentInfo(v)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			ownp, err := h.db.GetProfile(docinfo.OwnerUUID)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			updp, err := h.db.GetProfile(docinfo.UpdaterUUID)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			doclist = append(doclist, model.Document{
				Owner: model.Profile{
					UUID:    ownp.UUID,
					Name:    ownp.Name,
					IconURI: ownp.IconURI,
					Attr:    ownp.Attr,
					IsTeam:  (ownp.UUID[0] == 't'),
				},
				Updater: model.Profile{
					UUID:    updp.UUID,
					Name:    updp.Name,
					IconURI: updp.IconURI,
					Attr:    updp.Attr,
					IsTeam:  (updp.UUID[0] == 't'),
				},
				Title:      docinfo.Title,
				Permission: docinfo.Permission,
				CreatedAt:  docinfo.CreatedAt,
				UpdatedAt:  docinfo.UpdatedAt,
			})
		}
	}

	ret := model.FolderList{
		Folder:   follist,
		Document: doclist,
	}

	c.AbortWithStatusJSON(http.StatusOK, ret)
}

func (h *Handler) createFolderHandler(c *gin.Context) {
	parentfid := c.Param("folderid")
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var req model.CreateFolderReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	//TODO: permission check

	fid, err := h.db.CreateFolder(req.Name, req.Permission, parentfid, uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatusJSON(http.StatusOK, model.CreateFolderRes{FolderID: fid})
}

func (h *Handler) deleteFolderHandler(c *gin.Context) {
	fid := c.Param("folderid")
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//TODO: permission check
	if uuid == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err := h.db.DeleteFolder(fid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) moveFolderHandler(c *gin.Context) {
	fid := c.Param("folderid")
	targetfid := c.Param("targetfid")
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	//TODO: permission check
	if uuid == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err := h.db.MoveFolder(fid, targetfid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

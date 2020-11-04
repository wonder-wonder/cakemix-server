package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/model"
)

// FolderHandler is handlers of folders
func (h *Handler) FolderHandler(r *gin.RouterGroup) {
	folderck := r.Group("folder", h.CheckAuthMiddleware())
	folderck.GET("*folderid", h.getFolderHandler)
	folderck.POST(":folderid", h.createFolderHandler)
	folderck.DELETE(":folderid", h.deleteFolderHandler)
	folderck.PUT(":folderid/move/:targetfid", h.moveFolderHandler)
	folderck.PUT(":folderid", h.modifyFolderHandler)
}
func (h *Handler) getFolderHandler(c *gin.Context) {
	fid := c.Param("folderid")
	listtype := c.Query("type")

	follist := []model.Folder{}
	doclist := []model.Document{}

	fid = strings.TrimLeft(fid, "/")

	if fid == "" {
		var err error
		fid, err = h.db.GetRootFID()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	finfo, err := h.db.GetFolderInfo(fid)
	if err != nil {
		if err == db.ErrFolderNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !isRelatedUUID(c, finfo.OwnerUUID) && finfo.Permission == db.FilePermPrivate {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

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

			if !isRelatedUUID(c, folinfo.OwnerUUID) && folinfo.Permission == db.FilePermPrivate {
				continue
			}
			editable := isRelatedUUID(c, folinfo.OwnerUUID) || folinfo.Permission == db.FilePermReadWrite

			ownp, err := h.db.GetProfileByUUID(folinfo.OwnerUUID)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			updp, err := h.db.GetProfileByUUID(folinfo.UpdaterUUID)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			follist = append(follist, model.Folder{
				UUID: v,
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
				Permission: int(folinfo.Permission),
				CreatedAt:  folinfo.CreatedAt,
				UpdatedAt:  folinfo.UpdatedAt,
				Editable:   editable,
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

			if !isRelatedUUID(c, docinfo.OwnerUUID) && docinfo.Permission == db.FilePermPrivate {
				continue
			}
			editable := isRelatedUUID(c, docinfo.OwnerUUID) || docinfo.Permission == db.FilePermReadWrite

			ownp, err := h.db.GetProfileByUUID(docinfo.OwnerUUID)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			updp, err := h.db.GetProfileByUUID(docinfo.UpdaterUUID)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			doclist = append(doclist, model.Document{
				UUID: v,
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
				Permission: int(docinfo.Permission),
				CreatedAt:  docinfo.CreatedAt,
				UpdatedAt:  docinfo.UpdatedAt,
				Editable:   editable,
			})
		}
	}

	path, err := h.getPath(c, fid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ret := model.FolderList{
		Folder:   follist,
		Document: doclist,
		Path:     path,
	}

	c.AbortWithStatusJSON(http.StatusOK, ret)
}

func (h *Handler) createFolderHandler(c *gin.Context) {
	parentfid := c.Param("folderid")
	fname := c.Query("name")

	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

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

	owneruuid := uuid
	if isRelatedUUID(c, finfo.OwnerUUID) {
		owneruuid = finfo.OwnerUUID
	}

	fid, err := h.db.CreateFolder(fname, db.FilePermPrivate, parentfid, owneruuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = h.db.UpdateFolder(parentfid, uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, model.CreateFolderRes{FolderID: fid})
}

func (h *Handler) deleteFolderHandler(c *gin.Context) {
	fid := c.Param("folderid")

	finfo, err := h.db.GetFolderInfo(fid)
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

	docidlist, err := h.db.GetDocList(fid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	folidlist, err := h.db.GetFolderList(fid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if len(folidlist) > 0 || len(docidlist) > 0 {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	pfinfo, err := h.db.GetFolderInfo(finfo.ParentFolderUUID)
	if err != nil {
		if err == db.ErrFolderNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if !isRelatedUUID(c, pfinfo.OwnerUUID) && pfinfo.Permission != db.FilePermReadWrite {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = h.db.DeleteFolder(fid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = h.db.UpdateFolder(finfo.ParentFolderUUID, uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) moveFolderHandler(c *gin.Context) {
	fid := c.Param("folderid")
	targetfid := c.Param("targetfid")

	// Check folder permission
	finfo, err := h.db.GetFolderInfo(fid)
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

	sourcefid := finfo.ParentFolderUUID

	// Check original parent folder permission
	finfo, err = h.db.GetFolderInfo(sourcefid)
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

	err = h.db.MoveFolder(fid, targetfid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = h.db.UpdateFolder(fid, uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = h.db.UpdateFolder(sourcefid, uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = h.db.UpdateFolder(targetfid, uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) modifyFolderHandler(c *gin.Context) {
	fid := c.Param("folderid")

	// Check folder permission
	finfo, err := h.db.GetFolderInfo(fid)
	if err != nil {
		if err == db.ErrFolderNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	// Check owner
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	if finfo.OwnerUUID != uuid {
		teams, ok := getTeams(c)
		if !ok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		ng := true
		for _, v := range teams {
			if finfo.OwnerUUID == v {
				perm, err := h.db.GetTeamMemberPerm(v, uuid)
				if err != nil {
					c.AbortWithError(http.StatusInternalServerError, err)
				}
				if perm == db.TeamPermAdmin || perm == db.TeamPermOwner {
					ng = false
				}
				break
			}
		}
		if ng {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}

	req := model.FolderModifyReqModel{
		OwnerUUID:  finfo.OwnerUUID,
		Name:       finfo.Name,
		Permission: int(finfo.Permission),
	}

	err = c.BindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	finfo.OwnerUUID = req.OwnerUUID
	finfo.Name = req.Name
	finfo.Permission = db.FilePerm(req.Permission)

	err = h.db.UpdateFolderInfo(finfo)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) getPath(c *gin.Context, fid string) ([]model.Breadcrumb, error) {
	res := []model.Breadcrumb{}
	for {
		finfo, err := h.db.GetFolderInfo(fid)
		if err != nil {
			return res, err
		}
		if !isRelatedUUID(c, finfo.OwnerUUID) && finfo.Permission == db.FilePermPrivate {
			break
		}
		res = append([]model.Breadcrumb{{FolderID: fid, Title: finfo.Name}}, res...)
		fid = finfo.ParentFolderUUID
		if fid == "" {
			break
		}
	}
	return res, nil
}

package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/model"
	"github.com/wonder-wonder/cakemix-server/ot"
)

// DocumentHandler is handlers of documents
func (h *Handler) DocumentHandler(r *gin.RouterGroup) {
	r.GET("doc/:docid/ws", h.setJWTFromQuery(), h.CheckAuthMiddleware(), h.getOTHandler)
	docck := r.Group("doc", h.CheckAuthMiddleware())
	docck.GET(":docid", h.getDocumentHandler)
	docck.POST(":folderid", h.createDocumentHandler)
	docck.DELETE(":docid", h.deleteDocumentHandler)
	docck.PUT(":docid/move/:folderid", h.moveDocumentHandler)
	docck.PUT(":docid", h.modifyDocumentHandler)
}

func (h *Handler) getDocumentHandler(c *gin.Context) {
	did := c.Param("docid")

	if did == "" || did[0] != 'd' {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	dinfo, err := h.db.GetDocumentInfo(did)
	if err != nil {
		if err == db.ErrDocumentNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !isRelatedUUID(c, dinfo.OwnerUUID) && dinfo.Permission == db.FilePermPrivate {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	ownp, err := h.db.GetProfileByUUID(dinfo.OwnerUUID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	updp, err := h.db.GetProfileByUUID(dinfo.UpdaterUUID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	res := model.Document{
		UUID: did,
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
		Title:          dinfo.Title,
		Permission:     int(dinfo.Permission),
		CreatedAt:      dinfo.CreatedAt,
		UpdatedAt:      dinfo.UpdatedAt,
		Editable:       isRelatedUUID(c, dinfo.OwnerUUID) || dinfo.Permission == db.FilePermReadWrite,
		ParentFolderID: dinfo.ParentFolderUUID,
	}
	// doc, err := h.db.GetLatestDocument(did)
	// if err != nil {
	// 	c.AbortWithError(http.StatusInternalServerError, err)
	// 	return
	// }
	c.AbortWithStatusJSON(http.StatusOK, res)
}

func (h *Handler) createDocumentHandler(c *gin.Context) {
	parentfid := c.Param("folderid")

	if parentfid == "" || parentfid[0] != 'f' {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

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

	did, err := h.db.CreateDocument("Untitled", db.FilePermPrivate, parentfid, owneruuid, uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = h.db.UpdateFolder(parentfid, uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, model.CreateDocumentRes{DocumentID: did})
}

func (h *Handler) deleteDocumentHandler(c *gin.Context) {
	did := c.Param("docid")

	if did == "" || did[0] != 'd' {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	dinfo, err := h.db.GetDocumentInfo(did)
	if err != nil {
		if err == db.ErrDocumentNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	pfinfo, err := h.db.GetFolderInfo(dinfo.ParentFolderUUID)
	if err != nil {
		if err == db.ErrFolderNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Check document owner or parent folder owner
	if !isRelatedUUID(c, dinfo.OwnerUUID) && !isRelatedUUID(c, pfinfo.OwnerUUID) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Check parent folder permission
	if !isRelatedUUID(c, pfinfo.OwnerUUID) && pfinfo.Permission != db.FilePermReadWrite {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = h.db.DeleteDocument(did)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = h.db.UpdateFolder(dinfo.ParentFolderUUID, uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.AbortWithStatus(http.StatusOK)
}
func (h *Handler) moveDocumentHandler(c *gin.Context) {
	did := c.Param("docid")
	targetfid := c.Param("folderid")

	if did == "" || did[0] != 'd' {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if targetfid == "" || targetfid[0] != 'f' {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	dinfo, err := h.db.GetDocumentInfo(did)
	if err != nil {
		if err == db.ErrDocumentNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	sourcefid := dinfo.ParentFolderUUID

	pfinfo, err := h.db.GetFolderInfo(sourcefid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Check document owner or parent folder owner
	if !isRelatedUUID(c, dinfo.OwnerUUID) && !isRelatedUUID(c, pfinfo.OwnerUUID) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Check original parent folder permission
	if !isRelatedUUID(c, pfinfo.OwnerUUID) && pfinfo.Permission != db.FilePermReadWrite {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Check target folder permission
	tfinfo, err := h.db.GetFolderInfo(targetfid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !isRelatedUUID(c, tfinfo.OwnerUUID) && tfinfo.Permission != db.FilePermReadWrite {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = h.db.MoveDocument(did, targetfid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = h.db.UpdateDocument(did, uuid)
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

func (h *Handler) setJWTFromQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Request.Header.Set("Authorization", "Bearer "+token)
	}
}

func (h *Handler) modifyDocumentHandler(c *gin.Context) {
	did := c.Param("docid")

	if did == "" || did[0] != 'd' {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Check document permission
	dinfo, err := h.db.GetDocumentInfo(did)
	if err != nil {
		if err == db.ErrDocumentNotFound {
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
	if dinfo.OwnerUUID != uuid {
		teams, ok := getTeams(c)
		if !ok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		ng := true
		for _, v := range teams {
			if dinfo.OwnerUUID == v {
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
	req := model.DocumentModifyReqModel{
		OwnerUUID:  dinfo.OwnerUUID,
		Permission: int(dinfo.Permission),
	}

	err = c.BindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	dinfo.OwnerUUID = req.OwnerUUID
	dinfo.Permission = db.FilePerm(req.Permission)

	err = h.db.UpdateDocumentInfo(dinfo)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = h.db.UpdateDocument(did, uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) getOTHandler(c *gin.Context) {
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	docID := c.Param("docid")

	if docID == "" || docID[0] != 'd' {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Check permission
	docInfo, err := h.db.GetDocumentInfo(docID)
	if err != nil {
		if err == db.ErrDocumentNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !isRelatedUUID(c, docInfo.OwnerUUID) && docInfo.Permission == db.FilePermPrivate {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	editable := isRelatedUUID(c, docInfo.OwnerUUID) || docInfo.Permission == db.FilePermReadWrite

	// Setup websocket
	var wsupgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			//TODO
			return true
		},
	}
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v\n", err)
		return
	}
	defer conn.Close()

	p, err := h.db.GetProfileByUUID(uuid)
	if err != nil {
		log.Printf("OT handler error: %v", err)
		return
	}

	cl, err := ot.NewClient(conn, ot.ClientProfile{
		UUID:    uuid,
		Name:    p.Name,
		IconURI: p.IconURI,
	}, !editable)
	if err != nil {
		log.Printf("OT handler error: %v", err)
		return
	}
	h.otmgr.ClientConnect(cl, docID)
	cl.Loop()
}

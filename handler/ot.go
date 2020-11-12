package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/ot"
)

func (h *Handler) getOTHandler(c *gin.Context) {
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Check permission
	docID := c.Param("docid")
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
	name := p.Name

	sess, err := ot.OpenSession(h.db, docID)
	if err != nil {
		log.Printf("OT handler error: %v", err)
		return
	}
	otc := ot.NewOTClient(conn, uuid, name, editable)
	defer otc.Close()
	sess.AddClient(otc)
	sess.Request(ot.WSMsgTypeDoc, otc.ClientID, nil)

	otc.ClientLoop()
}

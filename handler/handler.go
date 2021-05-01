package handler

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/ot"
)

const (
	// ImageDir image dir path
	ImageDir = "/img"
)

var (
	dataDir                = ""
	mailTmplResetPW        = ""
	mailTmplRegist         = ""
	corsFrontHost          = "*" // CORS front host
	permitUserToCreateTeam = false
)

// Handler is object for handler function
type Handler struct {
	db    *db.DB
	otmgr *ot.Manager
}

type HandlerConf struct {
	DataDir                string
	MailTemplateResetPW    string
	MailTemplateRegist     string
	CORSHost               string
	PermitUserToCreateTeam bool
}

// NewHandler generates new Handler instance
func NewHandler(db *db.DB, conf HandlerConf) *Handler {
	dataDir = conf.DataDir
	// Init data dir
	err := os.MkdirAll(dataDir, 0700)
	if err != nil {
		panic("Directory init error:" + dataDir)
	}
	err = os.MkdirAll(path.Join(dataDir, ImageDir), 0700)
	if err != nil {
		panic("Directory init error:" + path.Join(dataDir, ImageDir))
	}
	if conf.MailTemplateResetPW == "" {
		panic("Mail template is not specified")
	}
	mailTmplResetPW = conf.MailTemplateResetPW
	if conf.MailTemplateRegist == "" {
		panic("Mail template is not specified")
	}
	mailTmplRegist = conf.MailTemplateRegist
	if conf.CORSHost != "" {
		corsFrontHost = conf.CORSHost
	}
	permitUserToCreateTeam = conf.PermitUserToCreateTeam
	otmgr, err := ot.NewManager(db)
	if err != nil {
		panic(err)
	}
	go otmgr.Loop()
	return &Handler{db: db, otmgr: otmgr}
}

// StopOTManager send stop request to OT manager
func (h *Handler) StopOTManager() {
	h.otmgr.StopOTManager()
}

func (h *Handler) notimplHandler(c *gin.Context) {
	c.String(http.StatusNotImplemented, "Not implemented: "+c.FullPath())
}

// CheckAuthMiddleware generates middleware to check JWT and set UUID
func (h *Handler) CheckAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header.Get("Authorization")
		hs := strings.SplitN(header, " ", 2)
		if len(hs) != 2 || hs[0] != "Bearer" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		uuid, sessionid, err := h.db.VerifyToken(hs[1])
		if err == db.ErrInvalidToken {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		} else if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		err = h.db.UpdateSessionLastUsed(uuid, sessionid)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Set("UUID", uuid)
		c.Set("SessionID", sessionid)
		teams, err := h.db.GetTeamsByUser(uuid)
		if err != nil {
			return
		}
		c.Set("Teams", teams)
	}
}

func getUUID(c *gin.Context) (string, bool) {
	dat, ok := c.Get("UUID")
	if !ok {
		return "", false
	}
	uuid, ok := dat.(string)
	return uuid, ok
}
func getSessionID(c *gin.Context) (string, bool) {
	dat, ok := c.Get("SessionID")
	if !ok {
		return "", false
	}
	sessionID, ok := dat.(string)
	return sessionID, ok
}
func getTeams(c *gin.Context) ([]string, bool) {
	dat, ok := c.Get("Teams")
	if !ok {
		return []string{}, false
	}
	uuid, ok := dat.([]string)
	return uuid, ok
}

func isRelatedUUID(c *gin.Context, uuid string) bool {
	myuuid, ok := getUUID(c)
	if !ok {
		return false
	}
	if uuid == myuuid {
		return true
	}
	teamuuids, ok := getTeams(c)
	if !ok {
		return false
	}
	for _, v := range teamuuids {
		if uuid == v {
			return true
		}
	}
	return false
}

// CORS supports cross origin resource sharing.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For preflight
		if c.Request.Method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", corsFrontHost)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, HEAD, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", c.Request.Header.Get("Access-Control-Request-Headers"))

			c.AbortWithStatus(http.StatusOK)
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", corsFrontHost)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, HEAD, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Origin, X-Csrftoken, Content-Type, Accept")
		c.Next()
	}
}

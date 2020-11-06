package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
)

const (
	corsFrontHost   = "*" // CORS front host
	imageDir        = "/img"
	profileImageDir = imageDir + "/prof"
	localImageBase  = "cakemix://"
)

var (
	dataDir = "./cmdat"
)

// Handler is object for handler function
type Handler struct {
	db *db.DB
}

// NewHandler generates new Handler instance
func NewHandler(db *db.DB) *Handler {
	if os.Getenv("DATADIR") != "" {
		dataDir = os.Getenv("DATADIR")
	}
	return &Handler{db: db}
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

		uuid, err := h.db.VerifyToken(hs[1])
		if err == db.ErrInvalidToken {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		} else if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Set("UUID", uuid)
		teams, err := h.db.GetTeamsByUser(uuid)
		if err != nil {
			return
		}
		c.Set("Teams", teams)
	}
}

// GetUUIDMiddleware generates middleware to set UUID if valid JWT is available
func (h *Handler) GetUUIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header.Get("Authorization")
		hs := strings.SplitN(header, " ", 2)
		if len(hs) != 2 || hs[0] != "Bearer" {
			return
		}

		uuid, err := h.db.VerifyToken(hs[1])
		if err != nil {
			return
		}
		c.Set("UUID", uuid)

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

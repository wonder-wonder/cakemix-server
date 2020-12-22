package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/handler"
)

func main() {
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	err := db.LoadKeys()
	if err != nil {
		panic(err)
	}

	db, err := db.OpenDB()
	if err != nil {
		panic(err)
	}

	// API handler
	r.Use(handler.CORS())
	v1 := r.Group("v1")
	v1Handler(v1, db)

	// Front serve
	FrontDir := ""
	if os.Getenv("FRONTDIR") != "" {
		FrontDir = os.Getenv("FRONTDIR")
	}
	if FrontDir != "" {
		r.Static("/dist", FrontDir)
		r.NoRoute(func(c *gin.Context) {
			if c.Request.URL.Path == "/dist/" {
				return
			}
			if strings.HasPrefix(c.Request.URL.Path, "/dist") {
				c.Request.URL.Path = "/dist/"
			} else {
				c.Request.URL.Path = "/dist" + c.Request.URL.Path
			}
			r.HandleContext(c)
		})
	}

	// Start web server
	fmt.Println("Start server")

	APIAddr := ""
	APIPort := "8081"
	if os.Getenv("APIADDR") != "" {
		APIAddr = os.Getenv("APIADDR")
	}
	if os.Getenv("PORT") != "" {
		APIPort = os.Getenv("PORT")
	}
	r.Run(APIAddr + ":" + APIPort)
}

func v1Handler(r *gin.RouterGroup, db *db.DB) {
	h := handler.NewHandler(db)
	h.AuthHandler(r)
	h.DocumentHandler(r)
	h.FolderHandler(r)
	h.ProfileHandler(r)
	h.TeamHandler(r)
	h.SearchHandler(r)
	h.ImageHandler(r)
}

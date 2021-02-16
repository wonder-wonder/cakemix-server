package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/handler"
	"github.com/wonder-wonder/cakemix-server/util"
)

func main() {
	util.LoadConfig()
	fileconf := util.GetFileConf()
	dbconf := util.GetDBConf()
	apiconf := util.GetAPIConf()
	mailconf := util.GetMailConf()

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	err := db.LoadKeys(fileconf.SignPrvKey, fileconf.SignPubKey)
	if err != nil {
		panic(err)
	}

	db, err := db.OpenDB(dbconf.Host, dbconf.Port, dbconf.User, dbconf.Pass, dbconf.Name)
	if err != nil {
		panic(err)
	}

	// API handler
	r.Use(handler.CORS())
	v1 := r.Group("v1")
	v1Handler(v1, db, fileconf.DataDir)

	// Front serve
	if fileconf.FrontDir != "" {
		r.Static("/dist", fileconf.FrontDir)
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

	// Init mail
	util.InitMail(mailconf.SendGridAPIKey, mailconf.FromAddr, mailconf.FromName)
	// DB cleaup
	go func() {
		time.Sleep(time.Minute)
		for {
			err := db.CleanupExpired()
			if err != nil {
				log.Printf("DB cleanup error: %v", err)
			}
			time.Sleep(time.Hour)
		}
	}()

	// Start web server
	fmt.Println("Start server")

	r.Run(apiconf.Host + ":" + apiconf.Port)
}

func v1Handler(r *gin.RouterGroup, db *db.DB, datadir string) {
	h := handler.NewHandler(db, datadir)
	h.AuthHandler(r)
	h.DocumentHandler(r)
	h.FolderHandler(r)
	h.ProfileHandler(r)
	h.TeamHandler(r)
	h.SearchHandler(r)
	h.ImageHandler(r)
}

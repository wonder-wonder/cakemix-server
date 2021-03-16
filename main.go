package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/handler"
	"github.com/wonder-wonder/cakemix-server/util"
)

var version string = "unknown version"

func main() {
	fmt.Printf("\nCakemix %s\n\n", version)
	if len(os.Args) > 1 {
		for i := 1; i < len(os.Args); i++ {
			switch strings.ToLower(os.Args[i]) {
			case "-c", "-conf":
				i++
				if i >= len(os.Args) {
					fmt.Fprintf(os.Stderr, "Option %s requires an argument\n", os.Args[i-1])
					os.Exit(1)
				}
				log.Printf("Loading config %s", os.Args[i])
				err := util.LoadConfigFile(os.Args[i])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error occured while loading config: %v\n", err)
					os.Exit(1)
				}
			default:
				fmt.Fprintf(os.Stderr, "Unknown option: %s\n", os.Args[i])
				os.Exit(1)
			}
		}
	}

	util.LoadConfigEnv()
	fileconf := util.GetFileConf()
	dbconf := util.GetDBConf()
	apiconf := util.GetAPIConf()
	mailconf := util.GetMailConf()

	gin.SetMode(gin.ReleaseMode)
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
	hconf := handler.HandlerConf{
		DataDir:             fileconf.DataDir,
		MailTemplateResetPW: mailconf.TmplResetPW,
		MailTemplateRegist:  mailconf.TmplRegist,
		CORSHost:            apiconf.CORS,
	}
	v1Handler(v1, db, hconf)

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
		// Wait DB startup
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
	log.Println("Start server")

	err = r.Run(apiconf.Host + ":" + apiconf.Port)
	if err != nil {
		panic(err)
	}
}

func v1Handler(r *gin.RouterGroup, db *db.DB, hconf handler.HandlerConf) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGQUIT)
	signal.Notify(sig, syscall.SIGTERM)
	h := handler.NewHandler(db, hconf)
	h.AuthHandler(r)
	h.DocumentHandler(r)
	h.FolderHandler(r)
	h.ProfileHandler(r)
	h.TeamHandler(r)
	h.SearchHandler(r)
	h.ImageHandler(r)
	go func() {
		<-sig
		h.StopOTManager()
		os.Exit(0)
	}()
}

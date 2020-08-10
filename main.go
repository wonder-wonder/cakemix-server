package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/handler"
)

func main() {
	r := gin.Default()
	db, err := db.OpenDB()
	if err != nil {
		panic(err)
	}

	// API handler
	r.GET("/", helloHandler)
	r.Use(handler.CORS())

	v1 := r.Group("v1")
	v1Handler(v1, db)

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
func helloHandler(c *gin.Context) {
	// c.Abort()
	c.String(http.StatusOK, "Hello, world!")
	c.Abort()
}

func v1Handler(r *gin.RouterGroup, db *db.DB) {
	h := handler.NewHandler(db)
	h.AuthHandler(r)
	h.DocumentHandler(r)
	h.FolderHandler(r)
	h.ProfileHandler(r)
	h.TeamHandler(r)
}

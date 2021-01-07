package handler

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
)

func testInit(tb testing.TB) (*gin.Engine, *db.DB, *gin.RouterGroup) {
	tb.Helper()
	os.Setenv("SIGNPRVKEY", "../signkey")
	os.Setenv("SIGNPUBKEY", "../signkey.pub")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	err := db.LoadKeys()
	if err != nil {
		tb.Errorf("testInit: %v", err)
	}
	db, err := db.OpenDB()
	if err != nil {
		tb.Errorf("testInit: %v", err)
	}
	v1 := r.Group("v1")
	return r, db, v1
}

package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/wonder-wonder/cakemix-server/db"
)

func testInit(tb testing.TB) *gin.Engine {
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
	h := NewHandler(db)
	h.AuthHandler(v1)
	h.DocumentHandler(v1)
	h.FolderHandler(v1)
	h.ProfileHandler(v1)
	h.TeamHandler(v1)
	h.SearchHandler(v1)
	h.ImageHandler(v1)

	return r
}

func testGetToken(tb testing.TB, r *gin.Engine) string {
	tb.Helper()

	reqbody := `{"id":"root","pass":"cakemix"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBufferString(reqbody))
	r.ServeHTTP(w, req)
	if !assert.Equal(tb, 200, w.Code) {
		tb.FailNow()
	}

	var res map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &res)
	if !assert.NoError(tb, err, "fail to umarshal json:\n%v", err) {
		tb.FailNow()
	}
	jwt := res["jwt"]
	if !assert.NotEmpty(tb, jwt) {
		tb.FailNow()
	}
	return jwt
}

func testOpenDB(tb testing.TB) *sql.DB {
	var (
		dbHost = "cakemixpg"
		dbPort = "5432"
		dbUser = "postgres"
		dbPass = "postgres"
		dbName = "cakemix"
	)

	if os.Getenv("DBHOST") != "" {
		dbHost = os.Getenv("DBHOST")
	}
	if os.Getenv("DBPORT") != "" {
		dbPort = os.Getenv("DBPORT")
	}
	if os.Getenv("DBUSER") != "" {
		dbUser = os.Getenv("DBUSER")
	}
	if os.Getenv("DBPASS") != "" {
		dbPass = os.Getenv("DBPASS")
	}
	if os.Getenv("DBNAME") != "" {
		dbName = os.Getenv("DBNAME")
	}
	db, err := sql.Open("postgres", "host= "+dbHost+" port="+dbPort+" user="+dbUser+" dbname="+dbName+" password="+dbPass+" sslmode=disable")
	assert.NoError(tb, err)
	return db
}

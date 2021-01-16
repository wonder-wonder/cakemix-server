package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/wonder-wonder/cakemix-server/db"
)

func TestMain(m *testing.M) {
	println("Prepare test data...")
	db, err := testOpenDB()
	if err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	dbexec := func(sql string) {
		_, err := tx.Exec(sql)
		if err != nil {
			if err2 := tx.Rollback(); err2 != nil {
				err = fmt.Errorf("%v: %v", err2, err)
			}
			panic(err)
		}
	}

	// Test user
	// Username: test1, Email: user1@example.com, Pass: pass
	dbexec("INSERT INTO username VALUES('urtsqctxpdg3ypzan','user1');")
	dbexec("INSERT INTO auth VALUES('urtsqctxpdg3ypzan','user1@example.com','4NoWTUFzUl9cllIfpxpp8MssVY8sYYfNpG3Y3dTMLewKqMTSGjiNInvKc0VEA8hUVIuPrIH1xXy3jI74vBAZ4A==','K7CNjNYcvH1XXq8R');")
	dbexec("INSERT INTO profile VALUES('urtsqctxpdg3ypzan','','',1610798538,'','ja');")
	dbexec("INSERT INTO folder VALUES('fw2ytzvb2y5qqpjfk','urtsqctxpdg3ypzan','fdahpbkboamdbgnua','user1',0,1610798538,1610798538,'urtsqctxpdg3ypzan');")

	err = tx.Commit()
	if err != nil {
		if err2 := tx.Rollback(); err2 != nil {
			err = fmt.Errorf("%v: %v", err2, err)
		}
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}

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

func testOpenDB() (*sql.DB, error) {
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
	return sql.Open("postgres", "host= "+dbHost+" port="+dbPort+" user="+dbUser+" dbname="+dbName+" password="+dbPass+" sslmode=disable")
}

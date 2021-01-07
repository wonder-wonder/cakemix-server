package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/wonder-wonder/cakemix-server/db"
)

func TestAuthHandler(t *testing.T) {
	os.Setenv("SIGNPRVKEY", "../signkey")
	os.Setenv("SIGNPUBKEY", "../signkey.pub")

	r, db, v1 := testInit(t)

	h := NewHandler(db)
	h.AuthHandler(v1)
	token := ""

	t.Run("Login", func(t *testing.T) {
		type req struct {
			body string
		}
		type res struct {
			code int
			// body string
		}
		tests := []struct {
			name string
			req  req
			res  res
		}{
			{
				name: "Root",
				req: req{
					body: `{"id":"root","pass":"cakemix"}`,
				},
				res: res{
					code: 200,
					// body: "",
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/auth/login", bytes.NewBufferString(tt.req.body))
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.res.code, w.Code)

				resraw := w.Body.Bytes()
				if string(resraw) == "" {
					t.Fatalf("should be string, got empty string")
				}

				var res map[string]string
				err := json.Unmarshal(resraw, &res)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}
				jwt, ok := res["jwt"]
				if !assert.True(t, ok, "should has jwt, got:\n%v", res) {
					t.FailNow()
				}
				if !assert.NotEmpty(t, jwt) {
					t.FailNow()
				}
				token = jwt
				// var exp interface{}
				// json.Unmarshal([]byte(tt.res.body), exp)
				// assert.Equal(t, tt.res.body, w.Body.String())

			})
		}
	})
	t.Run("CheckToken", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			header map[string]string
		}
		type res struct {
			code int
		}
		tests := []struct {
			name string
			req  req
			res  res
		}{
			{
				name: "Root",
				req: req{
					header: map[string]string{"Authorization": `Bearer ` + token},
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/auth/check/token", nil)
				for hk, hv := range tt.req.header {
					req.Header.Set(hk, hv)
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.res.code, w.Code)
			})
		}
	})
	t.Run("RegistGenToken", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			header map[string]string
		}
		type res struct {
			code int
		}
		tests := []struct {
			name string
			req  req
			res  res
		}{
			{
				name: "Root",
				req: req{
					header: map[string]string{"Authorization": `Bearer ` + token},
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/auth/regist/gen/token", nil)
				for hk, hv := range tt.req.header {
					req.Header.Set(hk, hv)
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.res.code, w.Code)

				resraw := w.Body.Bytes()
				if !assert.NotEmpty(t, resraw, "should be string, got empty string") {
					t.FailNow()
				}

				var res map[string]string
				err := json.Unmarshal(resraw, &res)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}
				invtoken, ok := res["token"]
				if !assert.True(t, ok, "should has jwt, got:\n%v", res) {
					t.FailNow()
				}
				if !assert.NotEmpty(t, invtoken) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("PassChange", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			header map[string]string
			body   string
		}
		type res struct {
			code int
		}
		tests := []struct {
			name string
			req  req
			res  res
		}{
			{
				name: "Root",
				req: req{
					header: map[string]string{"Authorization": `Bearer ` + token},
					body:   `{"oldpass":"cakemix","newpass":"mixcake"}`,
				},
				res: res{
					code: 200,
				},
			},
			{
				name: "Root2",
				req: req{
					header: map[string]string{"Authorization": `Bearer ` + token},
					body:   `{"oldpass":"mixcake","newpass":"cakemix"}`,
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/auth/pass/change", bytes.NewBufferString(tt.req.body))
				for hk, hv := range tt.req.header {
					req.Header.Set(hk, hv)
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.res.code, w.Code)
			})
		}
	})
	t.Run("Logout", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			header map[string]string
		}
		type res struct {
			code int
		}
		tests := []struct {
			name string
			req  req
			res  res
		}{
			{
				name: "Root",
				req: req{
					header: map[string]string{"Authorization": `Bearer ` + token},
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/auth/logout", nil)
				for hk, hv := range tt.req.header {
					req.Header.Set(hk, hv)
				}
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.res.code, w.Code)
			})
		}
	})

	t.Run("CheckUserName", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("GetRegistPre", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("PostRegistPre", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("RegistVerify", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})

	t.Run("PassReset", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("GetPassResetVerify", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("PostPassResetVerify", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
}

func testInit(tb testing.TB) (*gin.Engine, *db.DB, *gin.RouterGroup) {
	tb.Helper()
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

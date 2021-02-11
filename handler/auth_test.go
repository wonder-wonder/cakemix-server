package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthHandler(t *testing.T) {
	r := testInit(t)
	db, err := testOpenDB()
	assert.NoError(t, err)
	token := ""
	invitetoken := ""

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
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}

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
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
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
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}

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
				invitetoken = invtoken
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
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("GetSession", func(t *testing.T) {
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
				req, _ := http.NewRequest("GET", "/v1/auth/session", nil)
				for hk, hv := range tt.req.header {
					req.Header.Set(hk, hv)
				}
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}

				resraw := w.Body.Bytes()
				if string(resraw) == "" {
					t.Fatalf("should be string, got empty string")
				}

				var res []interface{}
				err := json.Unmarshal(resraw, &res)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}
				if !assert.LessOrEqual(t, 1, len(res)) {
					t.FailNow()
				}

				res1, ok := res[0].(map[string]interface{})
				if !assert.True(t, ok, "res[0] should map[string]interface{}, got:\n%v", res) {
					t.FailNow()
				}

				attrs := []string{"sessionid", "lastlogin", "lastused", "ipaddr", "devinfo", "iscurrent"}
				for _, v := range attrs {
					_, ok := res1[v]
					if !assert.True(t, ok, "should has %s, got:\n%v", v, res) {
						t.FailNow()
					}
				}
			})
		}
	})
	t.Run("DeleteSession", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		jwtbase64 := strings.Split(token, ".")[1]
		jwtbase64len := len(jwtbase64)
		if len(jwtbase64)%4 > 0 {
			for i := 0; i < (4 - (jwtbase64len % 4)); i++ {
				jwtbase64 += "="
			}
		}
		jwtraw, err := base64.StdEncoding.DecodeString(jwtbase64)
		if !assert.NoError(t, err, "fail to decode base64:\n%v", err) {
			t.FailNow()
		}
		var res map[string]interface{}
		err = json.Unmarshal(jwtraw, &res)
		if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
			t.FailNow()
		}
		sessid, ok := res["jti"]
		if !assert.True(t, ok, "should has key 'jti', got:\n%v", res) {
			t.FailNow()
		}
		sessidstr, ok := res["jti"].(string)
		if !assert.True(t, ok, "should be string, got:\n%v", sessid) {
			t.FailNow()
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/v1/auth/session/"+sessidstr, nil)

		req.Header.Set("Authorization", `Bearer `+token)

		r.ServeHTTP(w, req)
		if !assert.Equal(t, 200, w.Code) {
			t.FailNow()
		}
		token = testGetToken(t, r)
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
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
			})
		}
	})

	t.Run("CheckUserName", func(t *testing.T) {
		if invitetoken == "" {
			t.SkipNow()
		}
		type req struct {
			username string
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
				name: "Test",
				req: req{
					username: "test",
				},
				res: res{
					code: 200,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/auth/check/user/"+tt.req.username+"/"+invitetoken, nil)
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("GetRegistPre", func(t *testing.T) {
		if invitetoken == "" {
			t.SkipNow()
		}
		type res struct {
			code int
			// body string
		}
		tests := []struct {
			name string
			res  res
		}{
			{
				name: "Normal",
				res: res{
					code: 200,
					// body: "",
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/auth/regist/pre/"+invitetoken, nil)
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("PostRegistPre", func(t *testing.T) {
		if invitetoken == "" {
			t.SkipNow()
		}
		type req struct {
			body string
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
				name: "Test",
				req: req{
					body: `{"email":"test@example.com","username":"test","password":"pass"}`,
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/auth/regist/pre/"+invitetoken, bytes.NewBufferString(tt.req.body))
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("RegistVerify", func(t *testing.T) {
		if invitetoken == "" {
			t.SkipNow()
		}
		type req struct {
			username string
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
				name: "Test",
				req: req{
					username: "test",
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				veritoken := ""
				dateint := time.Now().Unix()
				row := db.QueryRow("SELECT token FROM preuser WHERE username = $1 AND expdate > $2", tt.req.username, dateint)
				err := row.Scan(&veritoken)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.NotEmpty(t, veritoken) {
					t.FailNow()
				}

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/auth/regist/verify/"+veritoken, nil)
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
			})
		}
	})

	t.Run("PassReset", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			body string
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
				name: "Test",
				req: req{
					body: `{"email":"test@example.com"}`,
				},
				res: res{
					code: 200,
				},
			},
			{
				name: "Root",
				req: req{
					body: `{"email":"root@localhost"}`,
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/auth/pass/reset", bytes.NewBufferString(tt.req.body))
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("GetPassResetVerify", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			email string
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
				name: "Test",
				req: req{
					email: "test@example.com",
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				veritoken := ""
				dateint := time.Now().Unix()
				row := db.QueryRow("SELECT token FROM passreset,auth WHERE email = $1 AND expdate > $2 AND passreset.uuid = auth.uuid", tt.req.email, dateint)
				err := row.Scan(&veritoken)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.NotEmpty(t, veritoken) {
					t.FailNow()
				}

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/auth/pass/reset/verify/"+veritoken, nil)
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("PostPassResetVerify", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			email string
			body  string
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
				name: "Test",
				req: req{
					email: "test@example.com",
					body:  `{"newpass":"resetpass"}`,
				},
				res: res{
					code: 200,
				},
			},
			{
				name: "Root",
				req: req{
					email: "root@localhost",
					body:  `{"newpass":"cakemix"}`,
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				veritoken := ""
				dateint := time.Now().Unix()
				row := db.QueryRow("SELECT token FROM passreset,auth WHERE email = $1 AND expdate > $2 AND passreset.uuid = auth.uuid", tt.req.email, dateint)
				err := row.Scan(&veritoken)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.NotEmpty(t, veritoken) {
					t.FailNow()
				}

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/auth/pass/reset/verify/"+veritoken, bytes.NewBufferString(tt.req.body))
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
			})
		}
	})

	t.Run("GetLog", func(t *testing.T) {
		token = testGetToken(t, r)
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
				req, _ := http.NewRequest("GET", "/v1/auth/log", nil)
				for hk, hv := range tt.req.header {
					req.Header.Set(hk, hv)
				}
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}

				resraw := w.Body.Bytes()
				if string(resraw) == "" {
					t.Fatalf("should be string, got empty string")
				}

				var res map[string]interface{}
				err := json.Unmarshal(resraw, &res)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}

				attrs := []string{"offset", "len", "logs"}
				for _, v := range attrs {
					_, ok := res[v]
					if !assert.True(t, ok, "should has %s, got:\n%v", v, res) {
						t.FailNow()
					}
				}
				reslogsraw, ok := res["logs"].([]interface{})
				if !assert.True(t, ok, "should be array, got:\n%v", res) {
					t.FailNow()
				}
				if len(reslogsraw) > 0 {
					reslogs, ok := reslogsraw[0].(map[string]interface{})
					if !assert.True(t, ok, "should be map[string]interface{}, got:\n%v", res) {
						t.FailNow()
					}
					logsattrs := []string{"user", "date", "type"}
					for _, v := range logsattrs {
						_, ok := reslogs[v]
						if !assert.True(t, ok, "should has %s, got:\n%v", v, res) {
							t.FailNow()
						}
					}
				}
			})
		}
	})
}

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchHandler(t *testing.T) {
	r := testInit(t)
	token := testGetToken(t, r)

	t.Run("SearchUser", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			header map[string]string
			q      string
			limit  int
			offset int
		}
		type res struct {
			code   int
			maxlen int
		}
		tests := []struct {
			name string
			req  req
			res  res
		}{
			{
				name: "All",
				req: req{
					header: map[string]string{"Authorization": `Bearer ` + token},
					q:      "",
					limit:  1,
					offset: 0,
				},
				res: res{
					code:   200,
					maxlen: 1,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				param := ""
				if tt.req.q != "" {
					if param != "" {
						param += "&"
					}
					param += "q=" + tt.req.q
				}
				if tt.req.limit > 0 {
					if param != "" {
						param += "&"
					}
					param += "limit=" + strconv.Itoa(tt.req.limit)
					if tt.req.offset > 0 {
						if param != "" {
							param += "&"
						}
						param += "offset=" + strconv.Itoa(tt.req.offset)
					}
				}
				if param != "" {
					param = "?" + param
				}
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/search/user"+param, nil)
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

				var res map[string]interface{}
				err := json.Unmarshal(resraw, &res)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}
				total, ok := res["total"]
				if !assert.True(t, ok, "should has total, got:\n%v", res) {
					t.FailNow()
				}
				_, ok = total.(int)
				if !ok {
					// specifications of json parser
					_, ok = total.(float64)
				}
				if !assert.True(t, ok, "total should int, got:\n%v", reflect.TypeOf(total)) {
					t.FailNow()
				}
				users, ok := res["users"]
				if !assert.True(t, ok, "should has users, got:\n%v", res) {
					t.FailNow()
				}
				usersarr, ok := users.([]interface{})
				if !assert.True(t, ok, "total should array, got:\n%v", users) {
					t.FailNow()
				}
				if !assert.LessOrEqual(t, tt.res.maxlen, len(usersarr)) {
					t.FailNow()
				}
				fmt.Printf("%v\n", res)
			})
		}
	})
	t.Run("SearchTeam", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			header map[string]string
			q      string
			limit  int
			offset int
		}
		type res struct {
			code   int
			maxlen int
		}
		tests := []struct {
			name string
			req  req
			res  res
		}{
			{
				name: "All",
				req: req{
					header: map[string]string{"Authorization": `Bearer ` + token},
					q:      "",
					limit:  1,
					offset: 0,
				},
				res: res{
					code:   200,
					maxlen: 1,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				param := ""
				if tt.req.q != "" {
					if param != "" {
						param += "&"
					}
					param += "q=" + tt.req.q
				}
				if tt.req.limit > 0 {
					if param != "" {
						param += "&"
					}
					param += "limit=" + strconv.Itoa(tt.req.limit)
					if tt.req.offset > 0 {
						if param != "" {
							param += "&"
						}
						param += "offset=" + strconv.Itoa(tt.req.offset)
					}
				}
				if param != "" {
					param = "?" + param
				}
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/search/team"+param, nil)
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

				var res map[string]interface{}
				err := json.Unmarshal(resraw, &res)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}
				total, ok := res["total"]
				if !assert.True(t, ok, "should has total, got:\n%v", res) {
					t.FailNow()
				}
				_, ok = total.(int)
				if !ok {
					// specifications of json parser
					_, ok = total.(float64)
				}
				if !assert.True(t, ok, "total should int, got:\n%v", reflect.TypeOf(total)) {
					t.FailNow()
				}
				teams, ok := res["teams"]
				if !assert.True(t, ok, "should has teams, got:\n%v", res) {
					t.FailNow()
				}
				teamsarr, ok := teams.([]interface{})
				if !assert.True(t, ok, "total should array, got:\n%v", teams) {
					t.FailNow()
				}
				if !assert.LessOrEqual(t, tt.res.maxlen, len(teamsarr)) {
					t.FailNow()
				}
				fmt.Printf("%v\n", res)
			})
		}
	})
}

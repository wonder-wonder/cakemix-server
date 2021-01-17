package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfileHandler(t *testing.T) {
	r := testInit(t)
	token := testGetToken(t, r)

	t.Run("GetProfile", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			header map[string]string
			uuid   string
		}
		type res struct {
			code int
			body string
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
					uuid:   "ujafzavrqkqthqe54",
				},
				res: res{
					code: 200,
					body: `{
						"uuid": "ujafzavrqkqthqe54",
						"name": "root",
						"bio": "",
						"icon_uri": "",
						"created_at": 1,
						"attr": "",
						"is_team": false,
						"teams": [
							{
								"uuid": "tqssoagvfvlg3mky2",
								"name": "admin",
								"bio": "",
								"icon_uri": "",
								"created_at": 0,
								"attr": "",
								"is_team": true,
								"teams": null,
								"lang": "",
								"is_admin": false
							}
						],
						"lang": "ja",
						"is_admin": true
					}`,
				}, // bio,created_at,teams,lang,is_admin in team are default values because some team info is omitted.
			},
			{
				name: "User1",
				req: req{
					header: map[string]string{"Authorization": `Bearer ` + token},
					uuid:   "urtsqctxpdg3ypzan",
				},
				res: res{
					code: 200,
					body: `{
						"uuid": "urtsqctxpdg3ypzan",
						"name": "user1",
						"bio": "user1bio",
						"icon_uri": "user1iconuri",
						"created_at": 1610798538,
						"attr": "user1attr",
						"is_team": false,
						"teams": [],
						"lang": "ja",
						"is_admin": false
					}`,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/profile/"+tt.req.uuid, nil)
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

				var res interface{}
				err := json.Unmarshal(resraw, &res)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}
				var exp interface{}
				err = json.Unmarshal([]byte(tt.res.body), &exp)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}
				assert.Equal(t, exp, res)
			})
		}
	})
	t.Run("UpdateProfile", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
}

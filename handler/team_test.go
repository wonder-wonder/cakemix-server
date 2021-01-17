package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTeamHandler(t *testing.T) {
	r := testInit(t)
	db, err := testOpenDB()
	assert.NoError(t, err)
	token := testGetToken(t, r)

	t.Run("CreateTeam", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}

		type req struct {
			header   map[string]string
			teamname string
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
				name: "TestTeam",
				req: req{
					header:   map[string]string{"Authorization": `Bearer ` + token},
					teamname: "testteam",
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/team?name="+tt.req.teamname, nil)
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
	t.Run("AddTeamMember", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}

		type req struct {
			header   map[string]string
			teamname string
			body     string
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
				name: "TestTeam",
				req: req{
					header:   map[string]string{"Authorization": `Bearer ` + token},
					teamname: "testteam",
					body:     `{"member":"urtsqctxpdg3ypzan","permission":0}`,
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				teamid := ""
				row := db.QueryRow("SELECT uuid FROM username WHERE username = $1", tt.req.teamname)
				err := row.Scan(&teamid)
				assert.NoError(t, err)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/team/"+teamid+"/member", bytes.NewBufferString(tt.req.body))
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
	t.Run("ModifyTeamMember", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("GetTeamMember", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("DeleteTeamMember", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("RemoveTeam", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
}

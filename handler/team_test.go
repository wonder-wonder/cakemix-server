package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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
					body:     `{"member":"urtsqctxpdg3ypzan","permission":1}`,
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
				req, _ := http.NewRequest("PUT", "/v1/team/"+teamid+"/member", bytes.NewBufferString(tt.req.body))
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
	t.Run("GetTeamMember", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}

		type req struct {
			header   map[string]string
			teamname string
			query    string
		}
		type res struct {
			code  int
			total int
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
					query:    ``,
				},
				res: res{
					code:  200,
					total: 2,
				},
			},
			{
				name: "TestTeam",
				req: req{
					header:   map[string]string{"Authorization": `Bearer ` + token},
					teamname: "testteam",
					query:    `uuid=ujafzavrqkqthqe54`,
				},
				res: res{
					code:  200,
					total: 1,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				teamid := ""
				row := db.QueryRow("SELECT uuid FROM username WHERE username = $1", tt.req.teamname)
				err := row.Scan(&teamid)
				assert.NoError(t, err)

				url := "/v1/team/" + teamid + "/member"
				if tt.req.query != "" {
					url += "?" + tt.req.query
				}
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", url, nil)
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
				err = json.Unmarshal(resraw, &res)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}
				total, ok := res["total"]
				if !assert.True(t, ok, "should has total, got:\n%v", res) {
					t.FailNow()
				}
				totalint, ok := total.(int)
				if !ok {
					// specifications of json parser
					totalfloat := 0.0
					totalfloat, ok = total.(float64)
					totalint = int(totalfloat)
				}
				if !assert.True(t, ok, "total should be int, got:\n%v", reflect.TypeOf(total)) {
					t.FailNow()
				}
				if !assert.Equal(t, tt.res.total, totalint) {
					t.FailNow()
				}
				members, ok := res["members"]
				if !assert.True(t, ok, "should has members, got:\n%v", res) {
					t.FailNow()
				}
				memarr, ok := members.([]interface{})
				if !assert.True(t, ok, "members is not array, got:\n%v", members) {
					t.FailNow()
				}
				if !assert.Equal(t, tt.res.total, len(memarr)) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("DeleteTeamMember", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}

		type req struct {
			header     map[string]string
			teamname   string
			memberuuid string
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
					header:     map[string]string{"Authorization": `Bearer ` + token},
					teamname:   "testteam",
					memberuuid: "urtsqctxpdg3ypzan",
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
				req, _ := http.NewRequest("DELETE", "/v1/team/"+teamid+"/member?uuid="+tt.req.memberuuid, nil)
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
	t.Run("RemoveTeam", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}

		type req struct {
			header     map[string]string
			teamname   string
			memberuuid string
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
				teamid := ""
				row := db.QueryRow("SELECT uuid FROM username WHERE username = $1", tt.req.teamname)
				err := row.Scan(&teamid)
				assert.NoError(t, err)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("DELETE", "/v1/team/"+teamid, nil)
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
}

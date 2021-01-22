package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFolderHandler(t *testing.T) {
	r := testInit(t)
	token := testGetToken(t, r)
	newfid := ""

	t.Run("GetFolder", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			header   map[string]string
			folderid string
			listtype string
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
					header:   map[string]string{"Authorization": `Bearer ` + token},
					folderid: "fwk6al7nyj4qdufaz",
					listtype: "",
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				url := "/v1/folder/" + tt.req.folderid
				if tt.req.listtype != "" {
					url += "?type=" + tt.req.listtype
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
				if !assert.NotEmpty(t, resraw, "should be string, got empty string") {
					t.FailNow()
				}

				var res map[string]interface{}
				err := json.Unmarshal(resraw, &res)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}

				path, ok := res["path"]
				if !assert.True(t, ok, "should has path, got:\n%v", res) {
					t.FailNow()
				}
				_, ok = path.([]interface{})
				if !assert.True(t, ok, "path should array, got:\n%v", path) {
					t.FailNow()
				}

				if tt.req.listtype != "folder" {
					folder, ok := res["folder"]
					if !assert.True(t, ok, "should has folder, got:\n%v", res) {
						t.FailNow()
					}
					_, ok = folder.([]interface{})
					if !assert.True(t, ok, "folder should array, got:\n%v", folder) {
						t.FailNow()
					}
				}

				if tt.req.listtype != "document" {
					document, ok := res["document"]
					if !assert.True(t, ok, "should has document, got:\n%v", res) {
						t.FailNow()
					}
					_, ok = document.([]interface{})
					if !assert.True(t, ok, "document should array, got:\n%v", document) {
						t.FailNow()
					}
				}
			})
		}
	})
	t.Run("CreateFolder", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			header   map[string]string
			folderid string
			name     string
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
					header:   map[string]string{"Authorization": `Bearer ` + token},
					folderid: "fwk6al7nyj4qdufaz",
					name:     "FolderTest",
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/folder/"+tt.req.folderid+"?name="+tt.req.name, nil)
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

				newfidraw, ok := res["folder_id"]
				if !assert.True(t, ok, "should has folder_id, got:\n%v", res) {
					t.FailNow()
				}
				newfid, ok = newfidraw.(string)
				if !assert.True(t, ok, "should be string, got:\n%v", res) {
					t.FailNow()
				}
				if !assert.NotEmpty(t, newfid) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("MoveFolder", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		if newfid == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("UpdateFolderInfo", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		if newfid == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
	t.Run("RemoveFolder", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		if newfid == "" {
			t.SkipNow()
		}
		// TODO: impl test
		t.Skip("Not implemented.")
	})
}

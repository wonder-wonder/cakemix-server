package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocumentHandler(t *testing.T) {
	r := testInit(t)
	token := testGetToken(t, r)
	newdid := ""

	t.Run("CreateDoc", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			header   map[string]string
			folderid string
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
				name: "TestDocument",
				req: req{
					header:   map[string]string{"Authorization": `Bearer ` + token},
					folderid: "fwk6al7nyj4qdufaz",
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/doc/"+tt.req.folderid, nil)
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

				newdidraw, ok := res["doc_id"]
				if !assert.True(t, ok, "should has doc_id, got:\n%v", res) {
					t.FailNow()
				}
				newdid, ok = newdidraw.(string)
				if !assert.True(t, ok, "should be string, got:\n%v", res) {
					t.FailNow()
				}
				if !assert.NotEmpty(t, newdid) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("GetDocInfo", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		if newdid == "" {
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
				name: "TestDocument",
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
				req, _ := http.NewRequest("GET", "/v1/doc/"+newdid, nil)
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

				attrs := []string{"owner", "updater", "title", "body", "permission", "created_at", "updated_at", "editable", "parentfolderid"}
				for _, v := range attrs {
					_, ok := res[v]
					if !assert.True(t, ok, "should has %s, got:\n%v", v, res) {
						t.FailNow()
					}
				}

				uuid, ok := res["uuid"]
				if !assert.True(t, ok, "should has uuid, got:\n%v", res) {
					t.FailNow()
				}
				uuidstr, ok := uuid.(string)
				if !assert.True(t, ok, "path should string, got:\n%v", uuid) {
					t.FailNow()
				}
				if !assert.Equal(t, newdid, uuidstr) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("UpdateDocInfo", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		if newdid == "" {
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
				name: "TestDocument",
				req: req{
					header: map[string]string{"Authorization": `Bearer ` + token},
					body:   `{"permission":1}`,
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("PUT", "/v1/doc/"+newdid, bytes.NewBufferString(tt.req.body))
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
	t.Run("MoveDoc", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		if newdid == "" {
			t.SkipNow()
		}
		type req struct {
			header    map[string]string
			targetfid string
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
					header:    map[string]string{"Authorization": `Bearer ` + token},
					targetfid: "fhfprvdljyczssis7",
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("PUT", "/v1/doc/"+newdid+"/move/"+tt.req.targetfid, nil)
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
	t.Run("DuplicateDoc", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		if newdid == "" {
			t.SkipNow()
		}
		type req struct {
			header    map[string]string
			targetfid string
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
					header:    map[string]string{"Authorization": `Bearer ` + token},
					targetfid: "fhfprvdljyczssis7",
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/doc/"+newdid+"/copy/"+tt.req.targetfid, nil)
				for hk, hv := range tt.req.header {
					req.Header.Set(hk, hv)
				}
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}

				resraw := w.Body.Bytes()
				var res map[string]interface{}
				err := json.Unmarshal(resraw, &res)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}

				dupdidraw, ok := res["doc_id"]
				if !assert.True(t, ok, "should has doc_id, got:\n%v", res) {
					t.FailNow()
				}
				dupdid, ok := dupdidraw.(string)
				if !assert.True(t, ok, "should be string, got:\n%v", res) {
					t.FailNow()
				}
				if !assert.NotEmpty(t, dupdid) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("GetDocByRevision", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		type req struct {
			header   map[string]string
			docid    string
			revision int
		}
		type res struct {
			code int
			text string
		}
		tests := []struct {
			name string
			req  req
			res  res
		}{
			{
				name: "TestDoc2",
				req: req{
					header:   map[string]string{"Authorization": `Bearer ` + token},
					docid:    "dzhkyo37b63qk3yj5",
					revision: 1,
				},
				res: res{
					code: 200,
					text: "This is a test.",
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/doc/"+tt.req.docid+"/rev/"+strconv.Itoa(tt.req.revision), nil)
				for hk, hv := range tt.req.header {
					req.Header.Set(hk, hv)
				}
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
				if !assert.Equal(t, tt.res.text, w.Body.String()) {
					t.FailNow()
				}
			})
		}
	})
	t.Run("RemoveDoc", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		if newdid == "" {
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
				name: "TestDocument",
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
				req, _ := http.NewRequest("DELETE", "/v1/doc/"+newdid, nil)
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
	// TODO: impl DocWebSocket test
	// t.Run("DocWebSocket", func(t *testing.T) {
	// 	if token == "" {
	// 		t.SkipNow()
	// 	}
	// 	if newdid == "" {
	// 		t.SkipNow()
	// 	}
	// 	// TODO: impl test
	// 	t.Skip("Not implemented.")
	// })
}

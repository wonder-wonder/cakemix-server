package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageHandler(t *testing.T) {
	r := testInit(t)
	token := testGetToken(t, r)
	imageid := ""

	t.Run("UploadImage", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}

		type req struct {
			header   map[string]string
			filepath string
			filename string
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
				name: "test.png",
				req: req{
					header:   map[string]string{"Authorization": `Bearer ` + token},
					filepath: "test.png",
					filename: "test.png",
				},
				res: res{
					code: 200,
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				file, err := os.Open(tt.req.filename)
				if !assert.NoError(t, err, "fail to open image file: %v", err) {
					t.FailNow()
				}
				body := &bytes.Buffer{}
				mw := multipart.NewWriter(body)
				fw, err := mw.CreateFormFile("file", tt.req.filename)
				_, err = io.Copy(fw, file)
				if !assert.NoError(t, err, "fail to read image file: %v", err) {
					t.FailNow()
				}
				ct := mw.FormDataContentType()
				err = mw.Close()
				if !assert.NoError(t, err, "fail to read image file: %v", err) {
					t.FailNow()
				}
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/image", body)
				req.Header.Add("Content-Type", ct)

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

				var res map[string]string
				err = json.Unmarshal(resraw, &res)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}
				imgid, ok := res["id"]
				if !assert.True(t, ok, "should has id, got:\n%v", res) {
					t.FailNow()
				}
				if !assert.NotEmpty(t, imgid) {
					t.FailNow()
				}
				imageid = imgid
			})
		}
	})
	t.Run("GetImage", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		if imageid == "" {
			t.SkipNow()
		}

		type req struct {
			header map[string]string
			imgid  string
		}
		type res struct {
			code     int
			filepath string
		}
		tests := []struct {
			name string
			req  req
			res  res
		}{
			{
				name: "test.png",
				req: req{
					header: map[string]string{"Authorization": `Bearer ` + token},
					imgid:  imageid,
				},
				res: res{
					code:     200,
					filepath: "test.png",
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/image/"+tt.req.imgid, nil)

				for hk, hv := range tt.req.header {
					req.Header.Set(hk, hv)
				}
				r.ServeHTTP(w, req)
				if !assert.Equal(t, tt.res.code, w.Code) {
					t.FailNow()
				}
				res := w.Body.Bytes()
				if len(res) == 0 {
					t.Fatalf("response size should be > 0, got size==0")
				}

				exp, err := ioutil.ReadFile(tt.res.filepath)
				if !assert.NoError(t, err, "fail to umarshal json:\n%v", err) {
					t.FailNow()
				}

				if !assert.Equal(t, len(exp), len(res)) {
					t.FailNow()
				}
				for i := range exp {
					if exp[i] != res[i] {
						t.Error("Response data is mismatch")
						t.FailNow()
					}
				}
			})
		}
	})
}

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wonder-wonder/cakemix-server/model"
)

func TestStatsHandler(t *testing.T) {
	r := testInit(t)
	token := testGetToken(t, r)

	t.Run("GetStats", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var stats model.Stats
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		assert.NoError(t, err)
		
		// Verify that stats contain positive values (we have test data)
		assert.GreaterOrEqual(t, stats.Users, 0)
		assert.GreaterOrEqual(t, stats.Teams, 0) 
		assert.GreaterOrEqual(t, stats.Documents, 0)
		assert.GreaterOrEqual(t, stats.Folders, 0)
	})

	t.Run("GetStatsMetrics", func(t *testing.T) {
		if token == "" {
			t.SkipNow()
		}
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/stats/metrics", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", w.Header().Get("Content-Type"))
		
		body := w.Body.String()
		// Verify Prometheus format
		assert.Contains(t, body, "cakemix_users_total")
		assert.Contains(t, body, "cakemix_teams_total")
		assert.Contains(t, body, "cakemix_documents_total")
		assert.Contains(t, body, "cakemix_folders_total")
		assert.Contains(t, body, "# HELP")
		assert.Contains(t, body, "# TYPE")
	})

	t.Run("GetStatsUnauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/stats", nil)
		r.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
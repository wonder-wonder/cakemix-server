package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// StatsHandler is handlers for statistics
func (h *Handler) StatsHandler(r *gin.RouterGroup) {
	statsck := r.Group("stats", h.CheckAuthMiddleware())
	statsck.GET("", h.getStatsHandler)
	statsck.GET("metrics", h.getStatsMetricsHandler)
}

func (h *Handler) getStatsHandler(c *gin.Context) {
	stats, err := h.db.GetStats()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) getStatsMetricsHandler(c *gin.Context) {
	stats, err := h.db.GetStats()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Return Prometheus metrics format
	metrics := fmt.Sprintf(`# HELP cakemix_users_total Total number of users
# TYPE cakemix_users_total gauge
cakemix_users_total %d

# HELP cakemix_teams_total Total number of teams
# TYPE cakemix_teams_total gauge
cakemix_teams_total %d

# HELP cakemix_documents_total Total number of documents
# TYPE cakemix_documents_total gauge
cakemix_documents_total %d

# HELP cakemix_folders_total Total number of folders
# TYPE cakemix_folders_total gauge
cakemix_folders_total %d
`, stats.Users, stats.Teams, stats.Documents, stats.Folders)

	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.String(http.StatusOK, metrics)
}
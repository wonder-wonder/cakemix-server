package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/model"
)

// SearchHandler is handlers for search
func (h *Handler) SearchHandler(r *gin.RouterGroup) {
	profck := r.Group("search", h.CheckAuthMiddleware())
	profck.GET("user", h.searchUserHandler)
	profck.GET("team", h.searchTeamHandler)
}

func (h *Handler) searchUserHandler(c *gin.Context) {
	res := []model.Profile{}
	var err error

	q := c.Query("q")
	filters := strings.Split(c.Query("filter"), " ")
	searchfilter := []int{}
	for _, v := range filters {
		switch strings.ToLower(v) {
		case "locked":
			searchfilter = append(searchfilter, db.SearchFilterLocked)
		case "unlocked":
			searchfilter = append(searchfilter, db.SearchFilterNotLocked)
		}
	}

	lim := -1
	offset := -1
	if c.Query("limit") != "" {
		lim, err = strconv.Atoi(c.Query("limit"))
		if err != nil || lim <= 0 {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}
	if c.Query("offset") != "" {
		offset, err = strconv.Atoi(c.Query("offset"))
		if err != nil || offset < 0 {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}

	count, list, err := h.db.SearchUser(q, lim, offset, searchfilter)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	for _, v := range list {
		uprof, err := h.db.GetProfileByUUID(v)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		prof := model.Profile{
			UUID:      uprof.UUID,
			Name:      uprof.Name,
			Bio:       uprof.Bio,
			IconURI:   uprof.IconURI,
			CreatedAt: uprof.CreateAt,
			Attr:      uprof.Attr,
			IsTeam:    false,
			Teams:     []model.Profile{},
			Lang:      uprof.Lang,
		}

		prof.IsAdmin, err = h.db.IsAdmin(uprof.UUID)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		prof.IsLock, err = h.db.IsUserLocked(uprof.UUID)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		teams, err := h.db.GetTeamsByUser(v)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		for _, t := range teams {
			tprof, err := h.db.GetProfileByUUID(t)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			prof.Teams = append(prof.Teams, model.Profile{
				UUID:    tprof.UUID,
				Name:    tprof.Name,
				IconURI: tprof.IconURI,
				Attr:    tprof.Attr,
				IsTeam:  true,
			})
		}
		res = append(res, prof)
	}

	c.JSON(http.StatusOK, model.SearchUserRes{Total: count, Users: res})
}

func (h *Handler) searchTeamHandler(c *gin.Context) {
	res := []model.Profile{}
	var err error
	q := c.Query("q")
	lim := -1
	offset := -1
	if c.Query("limit") != "" {
		lim, err = strconv.Atoi(c.Query("limit"))
		if err != nil || lim <= 0 {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}
	if c.Query("offset") != "" {
		offset, err = strconv.Atoi(c.Query("offset"))
		if err != nil || offset < 0 {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}

	count, list, err := h.db.SearchTeam(q, lim, offset)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	for _, v := range list {
		uprof, err := h.db.GetProfileByUUID(v)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		prof := model.Profile{
			UUID:      uprof.UUID,
			Name:      uprof.Name,
			Bio:       uprof.Bio,
			IconURI:   uprof.IconURI,
			CreatedAt: uprof.CreateAt,
			Attr:      uprof.Attr,
			IsTeam:    true,
			Teams:     []model.Profile{},
			Lang:      uprof.Lang,
		}

		res = append(res, prof)
	}

	c.JSON(http.StatusOK, model.SearchTeamRes{Total: count, Teams: res})
}

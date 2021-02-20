package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/model"
)

// TeamHandler is handlers of team
func (h *Handler) TeamHandler(r *gin.RouterGroup) {
	teamck := r.Group("team", h.CheckAuthMiddleware())
	teamck.GET(":teamid/member", h.getTeamMemberHandler)
	teamck.POST(":teamid/member", h.addTeamMemberHandler)
	teamck.PUT(":teamid/member", h.editTeamMemberHandler)
	teamck.DELETE(":teamid/member", h.deleteTeamMemberHandler)
	teamck.POST("", h.createTeamHandler)
	teamck.DELETE(":teamid", h.deleteTeamHandler)
}

func (h *Handler) createTeamHandler(c *gin.Context) {
	teamname := c.Query("name")
	useruuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	isadmin, err := h.db.IsAdmin(useruuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !isadmin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	_, err = h.db.CreateTeam(teamname, useruuid)
	if err == db.ErrExistUser {
		c.AbortWithStatus(http.StatusConflict)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) deleteTeamHandler(c *gin.Context) {
	teamuuid := c.Param("teamid")
	useruuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	perm, err := h.db.GetTeamMemberPerm(teamuuid, useruuid)
	if err == db.ErrUserNotFound {
		c.AbortWithStatus(http.StatusForbidden)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if perm != db.TeamPermOwner {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = h.db.DeleteTeam(teamuuid)
	if err == db.ErrUserTeamNotFound {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) getTeamMemberHandler(c *gin.Context) {
	teamid := c.Param("teamid")
	limit := -1
	offset := -1
	userid := c.Query("uuid")
	var err error
	if c.Query("limit") != "" {
		limit, err = strconv.Atoi(c.Query("limit"))
		if err != nil || limit <= 0 {
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

	total, mem, err := h.db.GetTeamMember(teamid, limit, offset, userid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if total == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	res := model.MemberInfoRes{Total: total, Members: []model.MemberInfo{}}

	for _, v := range mem {
		prof, err := h.db.GetProfileByUUID(v.UserUUID)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		res.Members = append(res.Members, model.MemberInfo{
			Member: model.Profile{
				UUID:    prof.UUID,
				Name:    prof.Name,
				IconURI: prof.IconURI,
				Attr:    prof.Attr,
				IsTeam:  prof.UUID[0] == 't',
			},
			Permission: int(v.Permission),
		})
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) addTeamMemberHandler(c *gin.Context) {
	teamid := c.Param("teamid")
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	perm, err := h.db.GetTeamMemberPerm(teamid, uuid)
	if err == db.ErrUserNotFound {
		c.AbortWithStatus(http.StatusForbidden)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !(perm == db.TeamPermOwner || perm == db.TeamPermAdmin) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	var req model.MemberInfoReq
	err = c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if req.Member[0] == 't' {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	_, err = h.db.GetProfileByUUID(req.Member)
	if err == db.ErrUserTeamNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	_, err = h.db.GetTeamMemberPerm(teamid, req.Member)
	if err == nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else if err != db.ErrUserNotFound {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	perm = db.TeamPerm(req.Permission)
	if perm == db.TeamPermOwner {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err = h.db.AddTeamMember(teamid, req.Member, perm)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) editTeamMemberHandler(c *gin.Context) {
	teamid := c.Param("teamid")
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	perm, err := h.db.GetTeamMemberPerm(teamid, uuid)
	if err == db.ErrUserNotFound {
		c.AbortWithStatus(http.StatusForbidden)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !(perm == db.TeamPermOwner || perm == db.TeamPermAdmin) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	var req model.MemberInfoReq
	err = c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	newperm := db.TeamPerm(req.Permission)

	oldperm, err := h.db.GetTeamMemberPerm(teamid, req.Member)
	if err == db.ErrUserNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if oldperm == db.TeamPermOwner {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if oldperm == newperm {
		c.AbortWithStatus(http.StatusOK)
		return
	}

	if newperm == db.TeamPermOwner {
		if perm != db.TeamPermOwner {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		err = h.db.ChangeTeamOwner(teamid, uuid, req.Member)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.AbortWithStatus(http.StatusOK)
		return
	}

	err = h.db.ModifyTeamMember(teamid, req.Member, newperm)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) deleteTeamMemberHandler(c *gin.Context) {
	teamid := c.Param("teamid")
	memberuuid := c.Query("uuid")
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	perm, err := h.db.GetTeamMemberPerm(teamid, uuid)
	if err == db.ErrUserNotFound {
		c.AbortWithStatus(http.StatusForbidden)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !(memberuuid == uuid || perm == db.TeamPermOwner || perm == db.TeamPermAdmin) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	oldperm, err := h.db.GetTeamMemberPerm(teamid, memberuuid)
	if err == db.ErrUserNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if oldperm == db.TeamPermOwner {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = h.db.DeleteTeamMember(teamid, memberuuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

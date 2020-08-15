package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/model"
)

// ProfileHandler is handlers of profile
func (h *Handler) ProfileHandler(r *gin.RouterGroup) {
	profck := r.Group("profile", h.CheckAuthMiddleware())
	profck.GET(":name", h.getProfileHandler)
	profck.PUT(":name", h.updateProfileHandler)
}

func (h *Handler) getProfileHandler(c *gin.Context) {
	var res model.Profile
	name := c.Param("name")
	p, err := h.db.GetProfile(name)
	if err == db.ErrUserTeamNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	res = model.Profile{
		UUID:      p.UUID,
		Name:      p.Name,
		Bio:       p.Bio,
		IconURI:   p.IconURI,
		CreatedAt: p.CreateAt,
		Attr:      p.Attr,
		Lang:      p.Lang,
		IsTeam:    (p.UUID[0] == 't'),
		Teams:     []model.Profile{},
	}

	teams, err := h.db.GetTeamsByUser(p.UUID)
	for _, v := range teams {
		prof, err := h.db.GetProfileByUUID(v)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		res.Teams = append(res.Teams, model.Profile{
			UUID:    prof.UUID,
			Name:    prof.Name,
			IconURI: prof.IconURI,
			Attr:    prof.Attr,
			IsTeam:  true,
		})
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) updateProfileHandler(c *gin.Context) {
	var req model.ProfileReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	name := c.Param("name")
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	profdat, err := h.db.GetProfile(name)
	if err == db.ErrUserTeamNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	//TODO: Edit permission check(temporary user update only)
	if profdat.UUID != uuid {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if req.Name != nil {
		profdat.Name = *req.Name
	}
	if req.Bio != nil {
		profdat.Bio = *req.Bio
	}
	if req.IconURI != nil {
		profdat.IconURI = *req.IconURI
	}
	if req.Lang != nil {
		profdat.Lang = *req.Lang
	}

	err = h.db.SetProfile(profdat)
	if err == db.ErrUserTeamNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.AbortWithStatus(http.StatusOK)
}

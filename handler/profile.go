package handler

import (
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"

	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/model"
)

// ProfileHandler is handlers of profile
func (h *Handler) ProfileHandler(r *gin.RouterGroup) {
	profck := r.Group("profile", h.CheckAuthMiddleware())
	profck.GET(":uuid", h.getProfileHandler)
	profck.PUT(":uuid", h.updateProfileHandler)
	profck.GET(":uuid/icon", h.notimplHandler)
	profck.POST(":uuid/icon", h.updateProfileIconHandler)
	profck.DELETE(":uuid/icon", h.notimplHandler)
}

func (h *Handler) getProfileHandler(c *gin.Context) {
	var res model.Profile
	uuid := c.Param("uuid")
	p, err := h.db.GetProfileByUUID(uuid)
	if err == db.ErrUserTeamNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	isadmin, err := h.db.IsAdmin(p.UUID)
	if err != nil {
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
		IsAdmin:   isadmin,
	}

	teams, _ := getTeams(c)
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
	trguuid := c.Param("uuid")
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	profdat, err := h.db.GetProfileByUUID(trguuid)
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

	// if req.Name != nil {
	// 	profdat.Name = *req.Name
	// }
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

func (h *Handler) updateProfileIconHandler(c *gin.Context) {
	trguuid := c.Param("uuid")
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	profdat, err := h.db.GetProfileByUUID(trguuid)
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

	imgid, err := db.GenerateID(db.IDTypeImageID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Upload the file to specific dst.
	err = c.SaveUploadedFile(file, path.Join(dataDir, profileImageDir, imgid))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	oldimg := profdat.IconURI
	profdat.IconURI = localImageBase + imgid

	err = h.db.SetProfile(profdat)
	if err == db.ErrUserTeamNotFound {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(oldimg) > len(localImageBase) && oldimg[:len(localImageBase)] == localImageBase {
		// Icon is in the system
		oldimgid := oldimg[len(localImageBase):]
		err = os.Remove(path.Join(dataDir, profileImageDir, oldimgid))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.AbortWithStatus(http.StatusOK)
}

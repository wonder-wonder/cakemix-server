package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/model"
	"github.com/wonder-wonder/cakemix-server/util"
)

// AuthHandler is handlers of auth
func (h *Handler) AuthHandler(r *gin.RouterGroup) {
	auth := r.Group("auth")
	auth.POST("login", h.loginHandler)
	auth.POST("regist/pre/:token", h.registHandler)
	auth.GET("regist/pre/:token", h.registTokenCheckHandler)
	auth.POST("regist/verify/:token", h.registVerifyHandler)
	auth.POST("pass/reset", h.passResetHandler)
	auth.GET("pass/reset/verify/:token", h.passResetTokenCheckHandler)
	auth.POST("pass/reset/verify/:token", h.passResetVerifyHandler)
	auth.GET("check/user/:name/:token", h.checkUserNameHandler)

	authck := auth.Group("", h.CheckAuthMiddleware())
	authck.POST("logout", h.logoutHandler)
	authck.GET("check/token", h.checkTokenHandler)
	authck.GET("regist/gen/token", h.registTokenGenerateHandler)
	authck.POST("pass/change", h.passChangeHandler)
	authck.GET("session", h.getSessionHandler)
	authck.DELETE("session/:id", h.removeSessionHandler)
	authck.GET("log", h.getLogHandler)
}

func (h *Handler) loginHandler(c *gin.Context) {
	var req model.AuthLoginReq
	var res model.AuthLoginRes
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	uuid, err := h.db.PasswordCheck(req.ID, req.Pass)
	if err == db.ErrIDPassInvalid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	skey, err := db.GenerateID(db.IDTypeSessionID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = h.db.AddSession(uuid, skey, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = h.db.AddLogLogin(uuid, skey, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res.JWT, err = db.GenerateJWT(uuid, skey)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, res)
}
func (h *Handler) checkTokenHandler(c *gin.Context) {
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) logoutHandler(c *gin.Context) {
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	sessid, ok := getSessionID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err := h.db.RemoveSession(uuid, sessid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) registTokenGenerateHandler(c *gin.Context) {
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	token, err := h.db.GenerateInviteToken(uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatusJSON(http.StatusOK, model.AuthRegistGenTokenReq{Token: token})
}
func (h *Handler) registTokenCheckHandler(c *gin.Context) {
	token := c.Param("token")
	err := h.db.CheckInviteToken(token)
	if err == db.ErrInvalidToken {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}
func (h *Handler) registHandler(c *gin.Context) {
	var req model.AuthRegistReq

	invtoken := c.Param("token")
	err := h.db.CheckInviteToken(invtoken)
	if err == db.ErrInvalidToken {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	token, err := h.db.PreRegistUser(req.UserName, req.Email, req.Password)
	if err == db.ErrExistUser {
		c.AbortWithStatus(http.StatusConflict)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = util.SendMailWithTemplate(req.Email, req.UserName, "Verify Email address", mailTmplRegist, map[string]string{"NAME": req.UserName, "TOKEN": token})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = h.db.DeleteInviteToken(invtoken)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) registVerifyHandler(c *gin.Context) {
	token := c.Param("token")
	err := h.db.RegistUser(token)
	if err == db.ErrInvalidToken {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) checkUserNameHandler(c *gin.Context) {
	username := c.Param("name")
	token := c.Param("token")
	err := h.db.CheckInviteToken(token)
	if err == db.ErrInvalidToken {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	_, err = h.db.GetProfileByUsername(username)
	if err == nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	} else if err != db.ErrUserTeamNotFound {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) passChangeHandler(c *gin.Context) {
	sessid, ok := getSessionID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var req model.AuthPassChangeReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = h.db.ChangePass(uuid, req.OldPass, req.NewPass)
	if err == db.ErrIDPassInvalid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = h.db.AddLogPassChange(uuid, sessid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) passResetHandler(c *gin.Context) {
	var req model.AuthPassResetReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	uuid, token, err := h.db.ResetPass(req.Email)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if uuid == "" {
		c.AbortWithStatus(http.StatusOK)
		return
	}
	prof, err := h.db.GetProfileByUUID(uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	islocked, err := h.db.IsUserLocked(uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if islocked {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = util.SendMailWithTemplate(req.Email, prof.Name, "Reset password", mailTmplResetPW, map[string]string{"NAME": prof.Name, "TOKEN": token})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}
func (h *Handler) passResetTokenCheckHandler(c *gin.Context) {
	token := c.Param("token")
	_, err := h.db.ResetPassTokenCheck(token)
	if err == db.ErrInvalidToken {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}
func (h *Handler) passResetVerifyHandler(c *gin.Context) {
	var req model.AuthPassChangeReq
	token := c.Param("token")
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	uuid, err := h.db.ResetPassVerify(token, req.NewPass)
	if err == db.ErrInvalidToken {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = h.db.AddLogPassReset(uuid, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) getSessionHandler(c *gin.Context) {
	res := []model.AuthSession{}

	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	sessid, ok := getSessionID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	sess, err := h.db.GetSession(uuid)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	for _, v := range sess {
		res = append(res, model.AuthSession{
			SessionID:  v.SessionID,
			LastLogin:  v.LoginDate,
			LastUsed:   v.LastDate,
			IPAddr:     v.IPAddr,
			DeviceInfo: v.DeviceData,
			IsCurrent:  v.SessionID == sessid,
		})
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) removeSessionHandler(c *gin.Context) {
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	token := c.Param("id")
	err := h.db.RemoveSession(uuid, token)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) getLogHandler(c *gin.Context) {
	var err error
	useruuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	teamuuid, ok := getTeams(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	targetidraw := c.Query("targetid")
	targetid := []string{}
	if targetidraw != "" {
		targetid = strings.Split(targetidraw, ",")
		for _, tv := range targetid {
			flag := false
			for _, v := range teamuuid {
				if tv == v {
					flag = true
					break
				}
			}
			if !flag {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}
	} else {
		targetid = teamuuid
	}
	offset := c.Query("offset")
	offsetint := 0
	if offset != "" {
		offsetint, err = strconv.Atoi(offset)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	limit := c.Query("limit")
	limitint := 0
	if limit != "" {
		limitint, err = strconv.Atoi(limit)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}
	ltyperaw := c.Query("type")
	ltype := []string{}
	if ltyperaw != "" {
		ltype = strings.Split(ltyperaw, ",")
	}

	logs, err := h.db.GetLogs(offsetint, limitint, useruuid, targetid, ltype)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	res := model.AuthLogRes{Offset: offsetint, Length: len(logs), Logs: []model.AuthLog{}}
	for _, l := range logs {
		reslog := model.AuthLog{Date: l.Date, Type: l.Type}
		resprof, err := h.db.GetProfileByUUID(l.UUID)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		reslog.User = model.Profile{UUID: resprof.UUID, Name: resprof.Name, IconURI: resprof.IconURI,
			Attr: resprof.Attr, IsTeam: resprof.UUID[0] == 't'}
		switch l.Type {
		case db.LogTypeAuthLogin:
			loginlog, err := h.db.GetLoginPassResetLog(l.ExtDataID)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			reslog.Data = model.AuthLogLogin{SessionID: l.SessionID, IPAddr: loginlog.IPAddr, DeviceInfo: loginlog.DeviceData}
		case db.LogTypeAuthPassReset:
			passresetlog, err := h.db.GetLoginPassResetLog(l.ExtDataID)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			reslog.Data = model.AuthLogPassReset{IPAddr: passresetlog.IPAddr, DeviceInfo: passresetlog.DeviceData}
		case db.LogTypeAuthPassChange:
			reslog.Data = model.AuthLogPassChange{SessionID: l.SessionID}
		}
		res.Logs = append(res.Logs, reslog)
	}

	c.AbortWithStatusJSON(http.StatusOK, res)
}

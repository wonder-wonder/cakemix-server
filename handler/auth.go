package handler

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/model"
)

// AuthHandler is handlers of auth
func (h *Handler) AuthHandler(r *gin.RouterGroup) {
	auth := r.Group("auth")
	auth.POST("login", h.loginHandler)
	auth.POST("regist/pre/:token", h.notimplHandler)
	auth.POST("regist/verify/:token", h.notimplHandler)
	auth.POST("pass/reset", h.passResetHandler)
	auth.GET("pass/reset/verify/:token", h.passResetTokenCheckHandler)
	auth.POST("pass/reset/verify/:token", h.passResetVerifyHandler)

	authck := auth.Group("/", h.CheckAuthMiddleware())
	authck.POST("logout", h.logoutHandler)
	authck.GET("check/token", h.checkTokenHandler)
	authck.GET("regist/gen/token", h.notimplHandler)
	authck.POST("pass/change", h.passChangeHandler)
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
	var claims jwt.StandardClaims

	header := c.Request.Header.Get("Authorization")
	hs := strings.SplitN(header, " ", 2)
	if len(hs) != 2 || hs[0] != "Bearer" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	_, err := jwt.ParseWithClaims(hs[1], &claims, nil)
	if (err.(*jwt.ValidationError).Errors & jwt.ValidationErrorMalformed) != 0 {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	err = h.db.RemoveSession(claims.Audience, claims.Id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) registHandler(c *gin.Context) {
	var req model.AuthRegistReq
	err := c.BindJSON(&req)
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
	//TODO:sendmail
	println(token)
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

func (h *Handler) passChangeHandler(c *gin.Context) {
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
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) passResetHandler(c *gin.Context) {
	var req model.AuthPassResetReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	token, err := h.db.ResetPass(req.Email)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	//TODO:sendmail
	println(token)
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
	err = h.db.ResetPassVerify(token, req.NewPass)
	if err == db.ErrInvalidToken {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/interfaces/http"
)

func AuthHandlers(r *gin.RouterGroup) {
	a := http.Auth{}
	auth := r.Group("auth")
	auth.POST("login", func(c *gin.Context) { a.Login(c) })
	auth.POST("regist/pre/:token", func(c *gin.Context) { a.RegistRequest(c) })
	auth.GET("regist/pre/:token", func(c *gin.Context) { a.InvitationTokenCheck(c) })
	auth.POST("regist/verify/:token", func(c *gin.Context) { a.RegistVerify(c) })
	auth.POST("pass/reset", func(c *gin.Context) { a.PassResetRequest(c) })
	auth.GET("pass/reset/verify/:token", func(c *gin.Context) { a.PassResetTokenCheck(c) })
	auth.POST("pass/reset/verify/:token", func(c *gin.Context) { a.PassResetVerify(c) })
	auth.GET("check/user/:name/:token", func(c *gin.Context) { a.CheckToken(c) })

	authck := auth.Group("", h.CheckAuthMiddleware())
	authck.POST("logout", func(c *gin.Context) { a.Logout(c) })
	authck.GET("check/token", func(c *gin.Context) { a.CheckToken(c) })
	authck.GET("regist/gen/token", func(c *gin.Context) { a.MakeInvitation(c) })
	authck.POST("pass/change", func(c *gin.Context) { a.PassChange(c) })
}

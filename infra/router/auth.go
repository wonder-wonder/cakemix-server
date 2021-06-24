package router

import (
	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/interfaces/http"
)

func AuthHandlers(r *gin.RouterGroup) {
	a := http.Auth{}
	auth := r.Group("auth")
	auth.POST("login", func(c *gin.Context) { a.Login(New(c)) })
	auth.POST("regist/pre/:token", func(c *gin.Context) { a.RegistRequest(New(c)) })
	auth.GET("regist/pre/:token", func(c *gin.Context) { a.InvitationTokenCheck(New(c)) })
	auth.POST("regist/verify/:token", func(c *gin.Context) { a.RegistVerify(New(c)) })
	auth.POST("pass/reset", func(c *gin.Context) { a.PassResetRequest(New(c)) })
	auth.GET("pass/reset/verify/:token", func(c *gin.Context) { a.PassResetTokenCheck(New(c)) })
	auth.POST("pass/reset/verify/:token", func(c *gin.Context) { a.PassResetVerify(New(c)) })
	auth.GET("check/user/:name/:token", func(c *gin.Context) { a.CheckToken(New(c)) })

	authck := auth.Group("", CheckAuthMiddleware())
	authck.POST("logout", func(c *gin.Context) { a.Logout(New(c)) })
	authck.GET("check/token", func(c *gin.Context) { a.CheckToken(New(c)) })
	authck.GET("regist/gen/token", func(c *gin.Context) { a.MakeInvitation(New(c)) })
	authck.POST("pass/change", func(c *gin.Context) { a.PassChange(New(c)) })
}

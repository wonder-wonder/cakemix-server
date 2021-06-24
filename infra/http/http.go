package http

import (
	"github.com/gin-gonic/gin"
	"github.com/wonder-wonder/cakemix-server/handler"
	"github.com/wonder-wonder/cakemix-server/interfaces/db"
	"github.com/wonder-wonder/cakemix-server/interfaces/http"
)

type Context struct {
	c *gin.Context
}
type Handler struct {
	auth *http.Auth
}

func NewHandler(r *gin.RouterGroup, db db.DB, hconf handler.HandlerConf) *Handler {
	return &Handler{
		auth: http.NewAuth(db),
	}
}

func New(c *gin.Context) *Context {
	return &Context{c: c}
}

func (c *Context) GetUUID() string {
	panic("TODO: impl")
}

func (c *Context) GetSessionID() string {
	panic("TODO: impl")
}
func (c *Context) Param(string) string {
	panic("TODO: impl")
}
func (c *Context) Bind(interface{}) error {
	panic("TODO: impl")
}
func (c *Context) Status(int) {
	panic("TODO: impl")
}
func (c *Context) JSON(int, interface{}) {
	panic("TODO: impl")
}

func CheckAuthMiddleware() gin.HandlerFunc {
	panic("TODO: impl")
	// return func(c *gin.Context) {
	// }
}

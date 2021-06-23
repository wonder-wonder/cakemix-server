package http

import "github.com/gin-gonic/gin"

type Context struct {
	c *gin.Context
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

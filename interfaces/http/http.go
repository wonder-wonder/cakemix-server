package http

// Handler func type
type Handler func(*Context)

// TODO
type Context interface {
	Param(string) string
	Bind(interface{}) error
	Status(int)
	JSON(int, interface{})

	GetUUID() string
	GetSessionID() string
}

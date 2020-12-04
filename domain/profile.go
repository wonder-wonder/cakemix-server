package domain

// Profile model
type Profile struct {
	UUID      string
	Bio       string
	IconURI   string
	CreatedAt int64
	Attr      string
	Lang      string
}

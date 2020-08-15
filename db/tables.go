package db

// Auth table model
type Auth struct {
	UUID     string
	Email    string
	Password string
	Salt     string
}

// Profile table model
type Profile struct {
	UUID     string
	Name     string
	Bio      string
	IconURI  string
	CreateAt int64
	Attr     string
	Lang     string
}

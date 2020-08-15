package db

// Auth table model
type Auth struct {
	UUID     string
	Email    string
	Password string
	Salt     string
}

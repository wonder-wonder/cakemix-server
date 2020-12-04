package domain

// User model
type User struct {
	UUID     string
	Username string
	Email    string
	Password string
	Salt     string
}

// Invitation model
type Invitation struct {
	FromUUID string
	Token    string
	ExpDate  int64
}

// PreUser model
type PreUser struct {
	Info    User
	Token   string
	ExpDate int64
}

// Session model
type Session struct {
	UUID       string
	SessionID  string
	LoginDate  int64
	LastDate   int64
	ExpDate    int64
	IPAddr     string
	DeviceData string
}

// PassReset model
type PassReset struct {
	UUID    string
	Token   string
	ExpDate int64
}

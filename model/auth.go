package model

// AuthRegistGenTokenReq model
type AuthRegistGenTokenReq struct {
	Token string `json:"token"`
}

//AuthRegistReq model
type AuthRegistReq struct {
	Email    string `json:"email"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

//AuthLoginReq model
type AuthLoginReq struct {
	ID   string `json:"id"`
	Pass string `json:"pass"`
}

//AuthLoginRes model
type AuthLoginRes struct {
	JWT string `json:"jwt"`
}

//AuthPassChangeReq model
type AuthPassChangeReq struct {
	OldPass string `json:"oldpass,omitempty"`
	NewPass string `json:"newpass"`
}

//AuthPassResetReq model
type AuthPassResetReq struct {
	Email string `json:"email"`
}

//AuthSession model
type AuthSession struct {
	SessionID  string `json:"sessionid"`
	LastLogin  int64  `json:"lastlogin"`
	LastUsed   int64  `json:"lastused"`
	IPAddr     string `json:"ipaddr"`
	DeviceInfo string `json:"devinfo"`
	IsCurrent  bool   `json:"iscurrent"`
}

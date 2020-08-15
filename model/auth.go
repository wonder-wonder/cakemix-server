package model

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

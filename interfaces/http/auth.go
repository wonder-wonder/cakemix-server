package http

import "github.com/wonder-wonder/cakemix-server/usecase"

// import "github.com/gin-gonic/gin"

// Auth handler structure
type Auth struct {
	userUC       usecase.User
	sessionUC    usecase.Session
	invitationUC usecase.Invitation
	preUserUC    usecase.PreUser
	passResetUC  usecase.PassReset
}

// Login handler
func (h *Auth) Login(c Context) {
	panic("TODO: impl")
}

// RegistRequest handler
func (h *Auth) RegistRequest(c Context) {
	panic("TODO: impl")
}

// InvitationTokenCheck handler
func (h *Auth) InvitationTokenCheck(c Context) {
	panic("TODO: impl")
}

// RegistVerify handler
func (h *Auth) RegistVerify(c Context) {
	panic("TODO: impl")
}

// PassResetRequest handler
func (h *Auth) PassResetRequest(c Context) {
	panic("TODO: impl")
}

// PassResetTokenCheck handler
func (h *Auth) PassResetTokenCheck(c Context) {
	panic("TODO: impl")
}

// PassResetVerify handler
func (h *Auth) PassResetVerify(c Context) {
	panic("TODO: impl")
}

// CheckUsername handler
func (h *Auth) CheckUsername(c Context) {
	panic("TODO: impl")
}

// Logout handler
func (h *Auth) Logout(c Context) {
	panic("TODO: impl")
}

// CheckToken handler
func (h *Auth) CheckToken(c Context) {
	panic("TODO: impl")
}

// MakeInvitation handler
func (h *Auth) MakeInvitation(c Context) {
	panic("TODO: impl")
}

// PassChange handler
func (h *Auth) PassChange(c Context) {
	panic("TODO: impl")
}

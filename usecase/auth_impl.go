package usecase

import (
	"crypto/sha512"
	"encoding/base64"
	"strings"

	"github.com/wonder-wonder/cakemix-server/domain"
)

func NewUser(userR UserRepo, seclogRepo SecurityLogRepo) *User {
	return &User{userRepo: userR, securityLogRepo: seclogRepo}
}

// Authenticate checks userid and password. If successful, returns UUID
func (uc *User) Authenticate(id, pass string) (uuid string, err error) {
	panic("TODO: impl")
	var user domain.User

	// Get user info
	if strings.Contains(id, "@") {
		user, err = uc.userRepo.FindByEmail(id)
		if err != nil {
			return
		}
	} else {
		user, err = uc.userRepo.FindByUsername(id)
		if err != nil {
			return
		}
	}

	// Check password
	hpass := passhash(pass, user.Salt)
	if hpass != user.Password {
		// TODO: ERROR
		return
	}

	// TODO: generate session Token

	// TODO: add login log

}

func passhash(pass, salt string) string {
	hp := sha512.Sum512([]byte(salt + pass))
	return base64.StdEncoding.EncodeToString(hp[:])
}

// ChangePass checks old pass and changes pass to new one.
func (uc *User) ChangePass(uuid, old, new string) error {
	panic("TODO: impl")
}

// func (uc *User) UsernameToUUID(username string) (bool, error) {

// }

func generateSessionID() string {
	panic("TODO: impl")
}

// New starts new session.
func (uc *Session) New(uuid, ipaddr, devdata string) (domain.Session, error) {
	panic("TODO: impl")
}

// Close stops the session.
func (uc *Session) Close(uuid, sessionID string) error {
	panic("TODO: impl")
}

// GetToken returns session token
func (uc *Session) GetToken(sess domain.Session) (string, error) {
	panic("TODO: impl")
}

// FromToken checks the session token and returns the session data.
func (uc *Session) FromToken(token string) (domain.Session, error) {
	panic("TODO: impl")
}

func generateInvitationToken() string {
	panic("TODO: impl")
}

// New makes invitation for new user
func (uc *Invitation) New(from string) (domain.Invitation, error) {
	panic("TODO: impl")
}

// Get returns invitation info corresponding token
func (uc *Invitation) Get(token string) (domain.Invitation, error) {
	panic("TODO: impl")
}

func generateEmailVerificationToken() string {
	panic("TODO: impl")
}

// Request checks invitation token and prepares new user and send verification mail
func (uc *PreUser) Request(token string, user domain.User) error {
	panic("TODO: impl")
}

// Register checks verification token and register the new user
func (uc *PreUser) Register(token string) error {
	panic("TODO: impl")
}

func generatePassResetVerificationToken() string {
	panic("TODO: impl")
}

// Request makes token and sends verification email
func (uc *PassReset) Request(email string) error {
	panic("TODO: impl")
}

// Get returns pass reset data corresponding token
func (uc *PassReset) Get(token string) (domain.PassReset, error) {
	panic("TODO: impl")
}

// Reset checks token and changes pass to new one.
func (uc *PassReset) Reset(token, newpass string) error {
	panic("TODO: impl")
}

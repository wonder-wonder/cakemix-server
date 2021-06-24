package usecase

import "github.com/wonder-wonder/cakemix-server/domain"

// UserRepo is interface of repository for user
type UserRepo interface {
	FindByEmail(string) (domain.User, error)
	FindByUsername(string) (domain.User, error)
	// FindByUUID(string) (domain.User, error)
	Add(domain.User) error
	Update(domain.User) error
}

// User data structure
type User struct {
	repo *UserRepo
}

func NewUser(repo UserRepo) *User {
	return &User{repo: &repo}
}

// Authorize checks userid and password. If successful, returns UUID
func (uc *User) Authorize(id, pass string) (string, error) {
	panic("TODO: impl")
}

func passhash(pass, salt string) string {
	panic("TODO: impl")
}

// ChangePass checks old pass and changes pass to new one.
func (uc *User) ChangePass(uuid, old, new string) error {
	panic("TODO: impl")
}

// func (uc *User) UsernameToUUID(username string) (bool, error) {

// }

// SessionRepo is interface of repository for session
type SessionRepo interface {
	Find(string, string) (domain.Session, error)
	Add(domain.Session) error
	Remove(string, string) error
}

// TokenGenerator is interface of token generator
type TokenGenerator interface {
	Generate(domain.Session) (string, error)
	Verify(string) error
}

// Session data structure
type Session struct {
	repo     SessionRepo
	tokengen TokenGenerator
}

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

// InvitationRepo is interface of repository for invitation.
type InvitationRepo interface {
	Add(domain.Invitation) error
	Find(string) (domain.Invitation, error)
	Remove(string) error
}

// Invitation data structure
type Invitation struct {
	repo InvitationRepo
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

// PreUserRepo is interface of repository for preuser
type PreUserRepo interface {
	Add(domain.PreUser) error
	FindByEmail(string) (domain.PreUser, error)
	FindByUsername(string) (domain.PreUser, error)
	Remove(string) error
}

// PreUser data structure
type PreUser struct {
	preUserRepo    PreUserRepo
	userRepo       UserRepo
	invitationRepo InvitationRepo
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

// PassResetRepo is interface of repository for pass reset
type PassResetRepo interface {
	Add(domain.PassReset) error
	Find(string) error
}

// PassReset data structure
type PassReset struct {
	userRepo      UserRepo
	passResetRepo PassResetRepo
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

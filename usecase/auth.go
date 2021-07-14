package usecase

import (
	"github.com/wonder-wonder/cakemix-server/domain"
)

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
	userRepo        UserRepo
	securityLogRepo SecurityLogRepo
}

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

type SecurityLogRepo interface {
	AddLogLogin(uuid string, sessionID string, ipaddr string, devinfo string) error
	AddLogPassReset(uuid string, ipaddr string, devinfo string) error
	AddLogPassChange(uuid string, ipaddr string, sessionID string) error
}

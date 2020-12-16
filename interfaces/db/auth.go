package db

import "github.com/wonder-wonder/cakemix-server/domain"

// UserRepo DB structure
type UserRepo struct {
	db DB
}

// FindByEmail returns user info corresponding email
func (r *UserRepo) FindByEmail(email string) (domain.User, error) {
	panic("TODO: impl")
}

// FindByUsername returns user info corresponding username
func (r *UserRepo) FindByUsername(username string) (domain.User, error) {
	panic("TODO: impl")
}

// Add adds user info
func (r *UserRepo) Add(user domain.User) error {
	panic("TODO: impl")
}

// Update updates user info
func (r *UserRepo) Update(user domain.User) error {
	panic("TODO: impl")
}

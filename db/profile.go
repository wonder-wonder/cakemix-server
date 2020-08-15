package db

import (
	"database/sql"
	"errors"
)

// Errors
var (
	ErrUserTeamNotFound = errors.New("User/team is not found")
)

// GetProfile returns the profile info
func (d *DB) GetProfile(name string) (Profile, error) {
	var p Profile
	r := d.db.QueryRow("SELECT p.uuid,p.name,p.bio,p.iconuri,p.createat,p.attr,p.lang FROM profile as p, username as u WHERE u.username = $1 AND p.uuid=u.uuid", name)
	err := r.Scan(&p.UUID, &p.Name, &p.Bio, &p.IconURI, &p.CreateAt, &p.Attr, &p.Lang)
	if err == sql.ErrNoRows {
		return p, ErrUserTeamNotFound
	} else if err != nil {
		return p, err
	}
	return p, nil
}

// SetProfile updates the profile info
func (d *DB) SetProfile(profile Profile) error {
	_, err := d.db.Exec("UPDATE profile SET name = $1, bio = $2, iconuri = $3, lang = $4 WHERE uuid = $5", profile.Name, profile.Bio, profile.IconURI, profile.Lang, profile.UUID)
	if err == sql.ErrNoRows {
		return ErrUserTeamNotFound
	} else if err != nil {
		return err
	}
	return nil
}

// GetProfileByUUID returns the profile info
func (d *DB) GetProfileByUUID(uuid string) (Profile, error) {
	var p Profile
	p.UUID = uuid
	r := d.db.QueryRow("SELECT name,bio,iconuri,createat,attr,lang FROM profile WHERE uuid=$1", uuid)
	err := r.Scan(&p.Name, &p.Bio, &p.IconURI, &p.CreateAt, &p.Attr, &p.Lang)
	if err == sql.ErrNoRows {
		return p, ErrUserTeamNotFound
	} else if err != nil {
		return p, err
	}
	return p, nil
}

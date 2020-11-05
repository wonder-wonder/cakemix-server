package db

import (
	"database/sql"
)

// GetProfileByUsername returns the profile info
func (d *DB) GetProfileByUsername(name string) (Profile, error) {
	var p Profile
	r := d.db.QueryRow("SELECT p.uuid,u.username,p.bio,p.iconuri,p.createat,p.attr,p.lang FROM profile as p, username as u WHERE u.username = $1 AND p.uuid=u.uuid", name)
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
	_, err := d.db.Exec("UPDATE profile SET bio = $1, iconuri = $2, lang = $3 WHERE uuid = $4", profile.Bio, profile.IconURI, profile.Lang, profile.UUID)
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
	r := d.db.QueryRow("SELECT username,bio,iconuri,createat,attr,lang FROM profile as p, username as u WHERE p.uuid = u.uuid AND p.uuid = $1", uuid)
	err := r.Scan(&p.Name, &p.Bio, &p.IconURI, &p.CreateAt, &p.Attr, &p.Lang)
	if err == sql.ErrNoRows {
		return p, ErrUserTeamNotFound
	} else if err != nil {
		return p, err
	}
	return p, nil
}

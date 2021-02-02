package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const teamNameAdmin = "admin"

// CreateTeam creates new team
func (d *DB) CreateTeam(teamname string, useruuid string) (string, error) {
	dateint := time.Now().Unix()
	teamuuid, err := GenerateID(IDTypeTeam)
	if err != nil {
		return "", err
	}

	uuid := ""
	//Check user username
	r := d.db.QueryRow("SELECT uuid FROM preuser WHERE username = $1 AND expdate > $2 UNION SELECT uuid FROM username WHERE username = $1", teamname, dateint)
	err = r.Scan(&uuid)
	if err == nil {
		return "", ErrExistUser
	} else if err != sql.ErrNoRows {
		return "", err
	}
	//Check ID duplication
	r = d.db.QueryRow("select uuid from username where uuid=$1", teamuuid)
	err = r.Scan(&uuid)
	if err == nil {
		return "", errors.New("Duplicate UUID is detected. You're so unlucky")
	} else if err != sql.ErrNoRows {
		return "", err
	}

	prof, err := d.GetProfileByUUID(useruuid)
	if err != nil {
		return "", err
	}

	fid, err := GenerateID(IDTypeFolder)
	if err != nil {
		return "", err
	}
	rootfid, err := d.GetRootFID()
	if err != nil {
		return "", err
	}

	tx, err := d.db.Begin()
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(`INSERT INTO username VALUES($1,$2)`, teamuuid, teamname)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return "", err
	}

	_, err = tx.Exec(`INSERT INTO profile VALUES($1,'','',$2,'',$3)`, teamuuid, dateint, prof.Lang)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return "", err
	}

	_, err = tx.Exec(`INSERT INTO teammember VALUES($1,$2,$3,$4)`, teamuuid, useruuid, TeamPermOwner, dateint)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return "", err
	}

	_, err = tx.Exec(`INSERT INTO folder VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, fid, teamuuid, rootfid, teamname, FilePermPrivate, dateint, dateint, useruuid)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return "", err
	}
	// //TODO: tags
	return teamuuid, nil
}

// DeleteTeam deletes team
func (d *DB) DeleteTeam(teamuuid string) error {
	cnt := 0
	//Check team exists
	if teamuuid[0] != 't' {
		return ErrUserTeamNotFound
	}
	r := d.db.QueryRow("SELECT count(*) FROM username WHERE uuid = $1 AND username != $2", teamuuid, teamNameAdmin)
	err := r.Scan(&cnt)
	if err != nil {
		return err
	}
	if cnt == 0 {
		return ErrUserTeamNotFound
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	owneruuid := ""
	r = tx.QueryRow("SELECT useruuid FROM teammember WHERE teamuuid = $1 AND permission = $2", teamuuid, TeamPermOwner)
	err = r.Scan(&owneruuid)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE folder SET owneruuid = $1 WHERE owneruuid = $2", owneruuid, teamuuid)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}

	_, err = tx.Exec("UPDATE document SET owneruuid = $1 WHERE owneruuid = $2", owneruuid, teamuuid)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}

	_, err = tx.Exec("DELETE FROM teammember WHERE teamuuid = $1", teamuuid)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}

	_, err = tx.Exec("DELETE FROM profile WHERE uuid = $1", teamuuid)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}

	_, err = tx.Exec("DELETE FROM username WHERE uuid = $1", teamuuid)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	//TODO: tags
	return nil
}

// GetTeamMember returns the member list of the team
func (d *DB) GetTeamMember(teamuuid string, limit int, offset int) (int, []TeamMember, error) {
	var res []TeamMember
	var count = 0

	r := d.db.QueryRow("SELECT COUNT(*) FROM teammember WHERE teamuuid = $1", teamuuid)
	err := r.Scan(&count)
	if err != nil {
		return 0, res, err
	}

	sql := "SELECT useruuid, permission FROM teammember WHERE teamuuid = $1 ORDER BY permission DESC, useruuid"
	param := []interface{}{teamuuid}
	if limit > 0 {
		param = append(param, limit)
		sql += " LIMIT $2"
		if offset > 0 {
			param = append(param, offset)
			sql += " OFFSET $3"
		}
	}

	rows, err := d.db.Query(sql, param...)
	if err != nil {
		return 0, res, err
	}
	defer rows.Close()
	for rows.Next() {
		var useruuid string
		var perm TeamPerm
		err = rows.Scan(&useruuid, &perm)
		if err != nil {
			return 0, res, err
		}
		res = append(res, TeamMember{TeamUUID: teamuuid, UserUUID: useruuid, Permission: perm})
	}
	if err = rows.Err(); err != nil {
		return 0, res, err
	}
	return count, res, nil
}

// GetTeamMemberPerm returns team member permission
func (d *DB) GetTeamMemberPerm(teamuuid string, useruuid string) (TeamPerm, error) {
	var perm TeamPerm
	row := d.db.QueryRow("SELECT permission FROM teammember WHERE teamuuid = $1 AND useruuid = $2", teamuuid, useruuid)
	err := row.Scan(&perm)
	if err == sql.ErrNoRows {
		return -1, ErrUserNotFound
	} else if err != nil {
		return -1, err
	}
	return perm, nil
}

// AddTeamMember adds new member into the team
func (d *DB) AddTeamMember(teamuuid string, useruuid string, perm TeamPerm) error {
	dateint := time.Now().Unix()
	_, err := d.db.Exec("INSERT INTO teammember VALUES($1,$2,$3,$4)", teamuuid, useruuid, perm, dateint)
	if err != nil {
		return err
	}
	return nil
}

// ModifyTeamMember updates team member permission
func (d *DB) ModifyTeamMember(teamuuid string, useruuid string, perm TeamPerm) error {
	_, err := d.db.Exec("UPDATE teammember SET permission = $1 WHERE teamuuid = $2 AND useruuid = $3", perm, teamuuid, useruuid)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTeamMember removes the member from the team
func (d *DB) DeleteTeamMember(teamuuid string, useruuid string) error {
	_, err := d.db.Exec("DELETE FROM teammember WHERE teamuuid = $1 AND useruuid = $2", teamuuid, useruuid)
	if err != nil {
		return err
	}
	return nil
}

// ChangeTeamOwner changes the owner of the team
func (d *DB) ChangeTeamOwner(teamuuid string, olduuid string, newuuid string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE teammember SET permission = $1 WHERE teamuuid = $2 AND useruuid = $3", TeamPermAdmin, teamuuid, olduuid)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	_, err = tx.Exec("UPDATE teammember SET permission = $1 WHERE teamuuid = $2 AND useruuid = $3", TeamPermOwner, teamuuid, newuuid)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	return nil
}

// GetTeamsByUser returns team IDs for specified user
func (d *DB) GetTeamsByUser(useruuid string) ([]string, error) {
	res := []string{}
	rows, err := d.db.Query("SELECT teamuuid FROM teammember WHERE useruuid = $1", useruuid)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		var teamuuid string
		err = rows.Scan(&teamuuid)
		if err != nil {
			return res, err
		}
		res = append(res, teamuuid)
	}
	if err = rows.Err(); err != nil {
		return res, err
	}
	return res, nil
}

// IsAdmin checks user is admin or not
func (d *DB) IsAdmin(uuid string) (bool, error) {
	cnt := 0
	r := d.db.QueryRow("SELECT COUNT(*) FROM teammember, username WHERE teamuuid = uuid AND useruuid = $1 AND username = $2", uuid, teamNameAdmin)
	err := r.Scan(&cnt)
	if err != nil {
		return false, err
	}

	return cnt == 1, nil
}

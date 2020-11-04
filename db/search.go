package db

import (
	"database/sql"
	"strconv"
)

// SearchUser returns the user uuid list of search
func (d *DB) SearchUser(query string, limit int, offset int) (int, []string, error) {
	var res []string
	var rows *sql.Rows
	var err error
	var count = 0
	if query != "" {
		r := d.db.QueryRow("SELECT COUNT(*) FROM auth as a, username as u WHERE a.uuid = u.uuid AND username like $1", query+"%")
		err = r.Scan(&count)
		if err != nil {
			return 0, res, err
		}
	} else {
		r := d.db.QueryRow("SELECT COUNT(*) FROM auth as a, username as u WHERE a.uuid = u.uuid")
		err = r.Scan(&count)
		if err != nil {
			return 0, res, err
		}
	}

	sql := "SELECT a.uuid FROM auth as a, username as u WHERE a.uuid = u.uuid"
	param := []interface{}{}
	if query != "" {
		param = append(param, query+"%")
		sql += " AND username like $" + strconv.Itoa(len(param))
	}

	if limit > 0 {
		param = append(param, limit)
		sql += " limit $" + strconv.Itoa(len(param))
		if offset > 0 {
			param = append(param, offset)
			sql += " offset $" + strconv.Itoa(len(param))
		}
	}
	rows, err = d.db.Query(sql, param...)
	if err != nil {
		return 0, res, err
	}
	defer rows.Close()
	for rows.Next() {
		var uuid string
		err = rows.Scan(&uuid)
		if err != nil {
			return 0, res, err
		}
		res = append(res, uuid)
	}
	if err = rows.Err(); err != nil {
		return 0, res, err
	}

	return count, res, nil
}

// SearchTeam returns the team uuid list of search
func (d *DB) SearchTeam(query string, limit int, offset int) (int, []string, error) {
	var res []string
	var rows *sql.Rows
	var err error

	var count = 0
	if query != "" {
		r := d.db.QueryRow("SELECT COUNT(*) FROM username WHERE uuid like 't%' AND username like $1", query+"%")
		err = r.Scan(&count)
		if err != nil {
			return 0, res, err
		}
	} else {
		r := d.db.QueryRow("SELECT COUNT(*) FROM username WHERE uuid like 't%'")
		err = r.Scan(&count)
		if err != nil {
			return 0, res, err
		}
	}

	sql := "SELECT uuid FROM username WHERE uuid like 't%'"
	param := []interface{}{}
	if query != "" {
		param = append(param, query+"%")
		sql += " AND username like $" + strconv.Itoa(len(param))
	}
	if limit > 0 {
		param = append(param, limit)
		sql += " limit $" + strconv.Itoa(len(param))
		if offset > 0 {
			param = append(param, offset)
			sql += " offset $" + strconv.Itoa(len(param))
		}
	}
	rows, err = d.db.Query(sql, param...)
	if err != nil {
		return 0, res, err
	}
	defer rows.Close()
	for rows.Next() {
		var uuid string
		err = rows.Scan(&uuid)
		if err != nil {
			return 0, res, err
		}
		res = append(res, uuid)
	}
	if err = rows.Err(); err != nil {
		return 0, res, err
	}

	return count, res, nil
}

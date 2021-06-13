package db

import (
	"database/sql"
	"fmt"
	"strconv"
)

const (
	SearchFilterLocked = iota
	SearchFilterNotLocked
)

// SearchUser returns the user uuid list of search
func (d *DB) SearchUser(query string, limit int, offset int, filters []int) (int, []string, error) {
	var res []string
	var rows *sql.Rows
	var err error
	var count = 0

	type sqlparam struct {
		col string
		op  string
		arg interface{}
	}
	sqlparams := []sqlparam{}
	if query != "" {
		sqlparams = append(sqlparams, sqlparam{col: "username", op: "ilike", arg: query + "%"})
	}
	for _, v := range filters {
		switch v {
		case SearchFilterLocked:
			sqlparams = append(sqlparams, sqlparam{col: "password", op: "like", arg: "$%"})
		case SearchFilterNotLocked:
			sqlparams = append(sqlparams, sqlparam{col: "password", op: "not like", arg: "$%"})
		}
	}

	sql := "SELECT COUNT(*) FROM auth INNER JOIN username ON auth.uuid = username.uuid"
	param := []interface{}{}
	if len(sqlparams) > 0 {
		sql += " WHERE"
		for i, v := range sqlparams {
			sql += fmt.Sprintf(" %s %s $%d", v.col, v.op, i+1)
			if i < len(sqlparams)-1 {
				sql += " AND"
			}
			param = append(param, v.arg)
		}
	}
	r := d.db.QueryRow(sql, param...)
	err = r.Scan(&count)
	if err != nil {
		return 0, res, err
	}

	sql = "SELECT auth.uuid FROM auth INNER JOIN username ON auth.uuid = username.uuid"
	param = []interface{}{}
	if len(sqlparams) > 0 {
		sql += " WHERE"
		for i, v := range sqlparams {
			sql += fmt.Sprintf(" %s %s $%d", v.col, v.op, i+1)
			if i < len(sqlparams)-1 {
				sql += " AND"
			}
			param = append(param, v.arg)
		}
	}

	sql += " ORDER BY username"
	if limit > 0 {
		param = append(param, limit)
		sql += fmt.Sprintf(" limit $%d", len(param))
		if offset > 0 {
			param = append(param, offset)
			sql += fmt.Sprintf(" offset $%d", len(param))
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
		r := d.db.QueryRow("SELECT COUNT(*) FROM username WHERE uuid like 't%' AND username ilike $1", query+"%")
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
		sql += " AND username ilike $" + strconv.Itoa(len(param))
	}
	sql += " ORDER BY username"
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

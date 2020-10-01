package db

import (
	"database/sql"
	"errors"
	"time"
)

var ErrFolderNotFound = errors.New("Folder is not found")

// GetFolderList returns the list of folders in the specified folder
func (d *DB) GetFolderList(fid string) ([]string, error) {
	var res []string
	rows, err := d.db.Query("SELECT uuid FROM folder WHERE parentfolderuuid = $1", fid)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		var uuid string
		err = rows.Scan(&uuid)
		if err != nil {
			return res, err
		}
		res = append(res, uuid)
	}
	if err = rows.Err(); err != nil {
		return res, err
	}
	return res, nil
}

// GetDocList returns the list of documents in the specified folder
func (d *DB) GetDocList(fid string) ([]string, error) {
	var res []string
	rows, err := d.db.Query("SELECT uuid FROM document WHERE parentfolderuuid = $1", fid)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		var uuid string
		err = rows.Scan(&uuid)
		if err != nil {
			return res, err
		}
		res = append(res, uuid)
	}
	if err = rows.Err(); err != nil {
		return res, err
	}
	return res, nil
}

// GetFolderInfo returns folder information
func (d *DB) GetFolderInfo(fid string) (Folder, error) {
	var ret Folder
	ret.UUID = fid
	r := d.db.QueryRow("SELECT owneruuid,parentfolderuuid,name,permission,createdat,updatedat,updateruuid FROM folder WHERE uuid = $1", ret.UUID)
	err := r.Scan(&ret.OwnerUUID, &ret.ParentFolderUUID, &ret.Name, &ret.Permission, &ret.CreatedAt, &ret.UpdatedAt, &ret.UpdaterUUID)
	if err == sql.ErrNoRows {
		return ret, ErrFolderNotFound
	} else if err != nil {
		return ret, err
	}
	return ret, nil
}

func (d *DB) CreateFolder(name string, permission int, parentfid string, owneruuid string) (string, error) {
	dateint := time.Now().Unix()
	fid, err := GenerateID(IDTypeFolder)
	if err != nil {
		return "", err
	}
	_, err = d.db.Exec(`INSERT INTO folder VALUES($1,$2,$3,$4,$4,$5,$6,$7,$8)`,
		fid, owneruuid, parentfid, name, permission, dateint, dateint, owneruuid)
	if err != nil {
		return "", err
	}
	return fid, nil
}

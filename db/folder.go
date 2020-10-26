package db

import (
	"database/sql"
	"time"
)

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

// CreateFolder creates new folder
func (d *DB) CreateFolder(name string, permission FilePerm, parentfid string, owneruuid string) (string, error) {
	dateint := time.Now().Unix()
	fid, err := GenerateID(IDTypeFolder)
	if err != nil {
		return "", err
	}
	_, err = d.db.Exec(`INSERT INTO folder VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
		fid, owneruuid, parentfid, name, permission, dateint, dateint, owneruuid)
	if err != nil {
		return "", err
	}
	return fid, nil
}

// DeleteFolder deletes folder
func (d *DB) DeleteFolder(fid string) error {
	_, err := d.db.Exec(`DELETE FROM folder WHERE uuid = $1`, fid)
	if err != nil {
		return err
	}
	return nil
}

// MoveFolder moves folder
func (d *DB) MoveFolder(fid string, targetfid string) error {
	_, err := d.db.Exec(`UPDATE folder SET parentfolderuuid = $1 WHERE uuid = $2`, targetfid, fid)
	if err != nil {
		return err
	}
	return nil
}

// GetRootFID returns root folder ID
func (d *DB) GetRootFID() (string, error) {
	fid := ""
	r := d.db.QueryRow("SELECT uuid FROM folder WHERE parentfolderuuid = ''")
	err := r.Scan(&fid)
	if err == sql.ErrNoRows {
		return fid, ErrFolderNotFound
	} else if err != nil {
		return fid, err
	}
	return fid, nil
}

// GetUserFID returns user folder ID
func (d *DB) GetUserFID() (string, error) {
	fid := ""

	rootfid, err := d.GetRootFID()
	if err != nil {
		return fid, err
	}

	r := d.db.QueryRow("SELECT uuid FROM folder WHERE parentfolderuuid = $1 AND name = 'User'", rootfid)
	err = r.Scan(&fid)
	if err == sql.ErrNoRows {
		return fid, ErrFolderNotFound
	} else if err != nil {
		return fid, err
	}
	return fid, nil
}

// UpdateFolderInfo modifies folder info
func (d *DB) UpdateFolderInfo(dat Folder) error {
	dateint := time.Now().Unix()
	_, err := d.db.Exec(`UPDATE folder SET owneruuid = $2, name = $3, permission = $4, updatedat = $5, updateruuid = $6 WHERE uuid = $1`,
		dat.UUID, dat.OwnerUUID, dat.Name, dat.Permission, dateint, dat.UpdaterUUID)
	if err != nil {
		return err
	}
	return nil
}

// UpdateFolder modifies folder update time
func (d *DB) UpdateFolder(fid string, updateruuid string) error {
	dateint := time.Now().Unix()
	_, err := d.db.Exec(`UPDATE folder SET updatedat = $2, updateruuid = $3 WHERE uuid = $1`,
		fid, dateint, updateruuid)
	if err != nil {
		return err
	}
	return nil
}

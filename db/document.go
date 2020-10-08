package db

import (
	"database/sql"
	"time"
)

// GetDocumentInfo returns document information
func (d *DB) GetDocumentInfo(fid string) (Document, error) {
	var ret Document
	ret.UUID = fid
	r := d.db.QueryRow("SELECT owneruuid,parentfolderuuid,title,permission,createdat,updatedat,updateruuid FROM document WHERE uuid = $1", ret.UUID)
	err := r.Scan(&ret.OwnerUUID, &ret.ParentFolderUUID, &ret.Title, &ret.Permission, &ret.CreatedAt, &ret.UpdatedAt, &ret.UpdaterUUID)
	if err == sql.ErrNoRows {
		return ret, ErrFolderNotFound
	} else if err != nil {
		return ret, err
	}
	return ret, nil
}

func (d *DB) CreateDocument(title string, permission FilePerm, parentfid string, owneruuid string) (string, error) {
	dateint := time.Now().Unix()
	did, err := GenerateID(IDTypeDocument)
	if err != nil {
		return "", err
	}
	_, err = d.db.Exec(`INSERT INTO document VALUES($1,$2,$3,$4,$4,$5,$6,$7,$8)`,
		did, owneruuid, parentfid, title, permission, dateint, dateint, owneruuid)
	if err != nil {
		return "", err
	}
	return did, nil
}

func (d *DB) DeleteDocument(did string) error {
	_, err := d.db.Exec(`DELETE FROM document WHERE uuid = $1`, did)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) MoveDocument(did string, targetfid string) error {
	_, err := d.db.Exec(`UPDATE document SET parentfolderuuid = $1 WHERE uuid = $2`, targetfid, did)
	if err != nil {
		return err
	}
	return nil
}

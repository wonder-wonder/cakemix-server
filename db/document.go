package db

import (
	"database/sql"
	"fmt"
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

// CreateDocument creates new document
func (d *DB) CreateDocument(title string, permission FilePerm, parentfid string, owneruuid string) (string, error) {
	dateint := time.Now().Unix()
	did, err := GenerateID(IDTypeDocument)
	if err != nil {
		return "", err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(`INSERT INTO document VALUES($1,$2,$3,$4,$4,$5,$6,$7,$8)`,
		did, owneruuid, parentfid, title, permission, dateint, dateint, owneruuid)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return "", err
	}
	_, err = tx.Exec(`INSERT INTO documentrevision VALUES($1,$2,$3)`,
		did, "", dateint)
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
	return did, nil
}

// DeleteDocument deletes document
func (d *DB) DeleteDocument(did string) error {
	_, err := d.db.Exec(`DELETE FROM document WHERE uuid = $1`, did)
	if err != nil {
		return err
	}
	return nil
}

// MoveDocument moves document to target folder
func (d *DB) MoveDocument(did string, targetfid string) error {
	_, err := d.db.Exec(`UPDATE document SET parentfolderuuid = $1 WHERE uuid = $2`, targetfid, did)
	if err != nil {
		return err
	}
	return nil
}

// GetLatestDocument returns document data
func (d *DB) GetLatestDocument(did string) (string, error) {
	text := ""
	r := d.db.QueryRow("SELECT text FROM document AS d,documentrevision AS dr WHERE dr.uuid = $1 AND dr.uuid = d.uuid AND d.updatedat = dr.updatedat", did)
	err := r.Scan(&text)
	if err == sql.ErrNoRows {
		return "", ErrFolderNotFound
	} else if err != nil {
		return "", err
	}
	return text, nil
}

// SaveDocument store the document data
func (d *DB) SaveDocument(did string, updateruuid string, text string) error {
	dateint := time.Now().Unix()
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE document SET updatedat = $1 WHERE uuid = $2`, dateint, did)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	_, err = tx.Exec(`INSERT INTO documentrevision VALUES($1,$2,$3)`,
		did, text, dateint)
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

// UpdateDocumentInfo modifies document info
func (d *DB) UpdateDocumentInfo(dat Document) error {
	dateint := time.Now().Unix()
	_, err := d.db.Exec(`UPDATE document SET owneruuid = $2, permission = $3, updateat = $4, updateruuid = $5 WHERE uuid = $1`,
		dat.UUID, dat.OwnerUUID, dat.Permission, dateint, dat.UpdaterUUID)
	if err != nil {
		return err
	}
	return nil
}

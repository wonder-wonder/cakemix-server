package db

import "database/sql"

// GetDocumentInfo returns document information
func (d *DB) GetDocumentInfo(fid string) (Document, error) {
	var ret Document
	ret.UUID = fid
	r := d.db.QueryRow("SELECT owneruuid,parentfolderuuid,title,permission,createdat,updatedat,updateruuid,tagid FROM document WHERE uuid = $1", ret.UUID)
	err := r.Scan(&ret.OwnerUUID, &ret.ParentFolderUUID, &ret.Title, &ret.Permission, &ret.CreatedAt, &ret.UpdatedAt, &ret.UpdaterUUID, &ret.TagID)
	if err == sql.ErrNoRows {
		return ret, ErrFolderNotFound
	} else if err != nil {
		return ret, err
	}
	return ret, nil
}

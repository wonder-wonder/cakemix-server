package db

import "github.com/wonder-wonder/cakemix-server/model"

// GetStats returns system statistics
func (d *DB) GetStats() (*model.Stats, error) {
	stats := &model.Stats{}
	var err error

	// Count users (UUIDs starting with 'u')
	err = d.db.QueryRow("SELECT COUNT(*) FROM username WHERE uuid LIKE 'u%'").Scan(&stats.Users)
	if err != nil {
		return nil, err
	}

	// Count teams (UUIDs starting with 't')
	err = d.db.QueryRow("SELECT COUNT(*) FROM username WHERE uuid LIKE 't%'").Scan(&stats.Teams)
	if err != nil {
		return nil, err
	}

	// Count documents
	err = d.db.QueryRow("SELECT COUNT(*) FROM document").Scan(&stats.Documents)
	if err != nil {
		return nil, err
	}

	// Count folders
	err = d.db.QueryRow("SELECT COUNT(*) FROM folder").Scan(&stats.Folders)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
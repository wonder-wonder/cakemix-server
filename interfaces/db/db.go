package db

// DB is interface for database
type DB interface {
	Begin() error
	Commit() error
	Query(string, ...interface{}) (*DBRows, error)
	QueryRow(string, ...interface{}) (*DBRow, error)
	Exec() error
}

// DBRows is interface for rows of result
type DBRows interface {
	Close() error
	Next() bool
	Scan(dest ...interface{}) error
}

// DBRow is interface for row of result
type DBRow interface {
	Scan(dest ...interface{}) error
}

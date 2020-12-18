package db

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq" //PostgreSQL driver
	"github.com/wonder-wonder/cakemix-server/interfaces/db"
)

var (
	dbHost = "cakemixpg"
	dbPort = "5432"
	dbUser = "postgres"
	dbPass = "postgres"
	dbName = "cakemix"
)

// DB holds DB connection
type DB struct {
	db *sql.DB
}

func initVars() {
	dbHost = os.Getenv("DBHOST")
	dbPort = os.Getenv("DBPORT")
	dbUser = os.Getenv("DBUSER")
	dbPass = os.Getenv("DBPASS")
	dbName = os.Getenv("DBNAME")
	if os.Getenv("DBHOST") != "" {
		dbHost = os.Getenv("DBHOST")
	}
	if os.Getenv("DBPORT") != "" {
		dbPort = os.Getenv("DBPORT")
	}
	if os.Getenv("DBUSER") != "" {
		dbUser = os.Getenv("DBUSER")
	}
	if os.Getenv("DBPASS") != "" {
		dbPass = os.Getenv("DBPASS")
	}
	if os.Getenv("DBNAME") != "" {
		dbName = os.Getenv("DBNAME")
	}
}

// OpenDB connects to DB server and return DB instance
func OpenDB() (db.DB, error) {
	initVars()
	db, err := sql.Open("postgres", "host= "+dbHost+" port="+dbPort+" user="+dbUser+" dbname="+dbName+" password="+dbPass+" sslmode=disable")
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

func (db *DB) Begin() error {
	panic("TODO: impl")
}
func (db *DB) Commit() error {
	panic("TODO: impl")

}
func (db *DB) Query(string, ...interface{}) (*db.DBRows, error) {
	panic("TODO: impl")

}
func (db *DB) QueryRow(string, ...interface{}) (*db.DBRow, error) {
	panic("TODO: impl")

}
func (db *DB) Exec() error {
	panic("TODO: impl")

}

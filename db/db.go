package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	_ "github.com/lib/pq" //PostgreSQL driver
)

// IDType is enum of types of ID
type IDType int

// IDType list
const (
	IDTypeVerifyToken IDType = iota
	IDTypeSessionID
	IDTypeSalt
	IDTypeImageID
	IDTypeUser
	IDTypeTeam
	IDTypeFolder
	IDTypeDocument
)

const (
	sizeVerifyToken = 24
	sizeSessionID   = 9
	sizeSalt        = 12
	sizeImageID     = 36
	sizeUser        = 10
	sizeProject     = 10
	sizeComment     = 10
)

// DB holds DB connection
type DB struct {
	db *sql.DB
}

// OpenDB connects to DB server and return DB instance
func OpenDB(dbHost, dbPort, dbUser, dbPass, dbName string) (*DB, error) {
	// initVars()
	db, err := sql.Open("postgres", "host= "+dbHost+" port="+dbPort+" user="+dbUser+" dbname="+dbName+" password="+dbPass+" sslmode=disable")
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

// GenerateID generates random IDs
func GenerateID(t IDType) (string, error) {
	size := 0
	var enc func([]byte) string
	switch t {
	case IDTypeVerifyToken:
		size = sizeVerifyToken
		enc = base64.URLEncoding.EncodeToString
	case IDTypeSessionID:
		size = sizeSessionID
		enc = base64.URLEncoding.EncodeToString
	case IDTypeSalt:
		size = sizeSalt
		enc = base64.StdEncoding.EncodeToString
	case IDTypeImageID:
		size = sizeImageID
		enc = base64.URLEncoding.EncodeToString
	case IDTypeUser:
		size = sizeUser
		enc = func(src []byte) string {
			return "u" + strings.ToLower(base32.StdEncoding.EncodeToString(src))
		}
	case IDTypeTeam:
		size = sizeUser
		enc = func(src []byte) string {
			return "t" + strings.ToLower(base32.StdEncoding.EncodeToString(src))
		}
	case IDTypeFolder:
		size = sizeProject
		enc = func(src []byte) string {
			return "f" + strings.ToLower(base32.StdEncoding.EncodeToString(src))
		}
	case IDTypeDocument:
		size = sizeProject
		enc = func(src []byte) string {
			return "d" + strings.ToLower(base32.StdEncoding.EncodeToString(src))
		}
	default:
		return "", errors.New("Unexpected IDType")
	}

	rd := make([]byte, size)
	_, err := rand.Read(rd)
	if err != nil {
		return "", err
	}
	return enc(rd), nil
}

// CleanupExpired removes expired record
func (d *DB) CleanupExpired() error {
	dateint := time.Now().Unix()

	_, err := d.db.Exec("DELETE FROM session WHERE expiredate < $1", dateint)
	if err != nil {
		return err
	}

	_, err = d.db.Exec("DELETE FROM invitetoken WHERE expdate < $1", dateint)
	if err != nil {
		return err
	}

	_, err = d.db.Exec("DELETE FROM preuser WHERE expdate < $1", dateint)
	if err != nil {
		return err
	}

	_, err = d.db.Exec("DELETE FROM passreset WHERE expdate < $1", dateint)
	if err != nil {
		return err
	}

	return nil
}

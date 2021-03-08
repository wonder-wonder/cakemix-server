package db

import (
	"crypto/rsa"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	loginSessionExpHours = 24 * 14
	verifyTokenExpHours  = 12
)

// #nosec G101
// LogTypeAuth enum
const (
	LogTypeAuthLogin      = "auth.login"
	LogTypeAuthPassReset  = "auth.passreset"
	LogTypeAuthPassChange = "auth.passchange"
)

var (
	signKey   *rsa.PrivateKey
	verifyKey *rsa.PublicKey
)

// LoadKeys read public/private keys
func LoadKeys(rsaPrivateKeyFile, rsaPublicKeyFile string) error {
	// #nosec G304
	// Signing (private) key
	signBytes, err := ioutil.ReadFile(rsaPrivateKeyFile)
	if err != nil {
		return err
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return err
	}

	// #nosec G304
	// Verification (public) key
	verifyBytes, err := ioutil.ReadFile(rsaPublicKeyFile)
	if err != nil {
		return err
	}
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return err
	}

	return nil
}

// PasswordCheck checks given ID and pass and returns UUID
func (d *DB) PasswordCheck(userid string, pass string) (string, error) {
	var auth Auth
	var r *sql.Row
	var sqlq = "SELECT a.uuid,a.email,a.password,a.salt FROM auth AS a,username AS u WHERE a.uuid = u.uuid AND u.username = $1" // userid is username
	if strings.Contains(userid, "@") {
		// userid is email addr
		sqlq = "SELECT uuid,email,password,salt FROM auth WHERE email = $1"
	}
	r = d.db.QueryRow(sqlq, userid)
	err := r.Scan(&auth.UUID, &auth.Email, &auth.Password, &auth.Salt)
	if err == sql.ErrNoRows {
		return "", ErrIDPassInvalid
	} else if err != nil {
		return "", err
	}

	hps := passhash(pass, auth.Salt)
	if hps == auth.Password {
		return auth.UUID, nil
	}
	return "", ErrIDPassInvalid
}

func passhash(pass string, salt string) string {
	hp := sha512.Sum512([]byte(salt + pass))
	return base64.StdEncoding.EncodeToString(hp[:])
}

// GenerateJWT generates JWT using UUID and sessionID
func GenerateJWT(uuid string, sessionid string) (string, error) {
	// create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.StandardClaims{
		Audience:  uuid,
		ExpiresAt: time.Now().Add(time.Hour * loginSessionExpHours).Unix(),
		Id:        sessionid,
	})

	tokenString, err := token.SignedString(signKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GenerateSessionID generates randomID for sessions
func GenerateSessionID() (string, error) {
	return GenerateID(IDTypeSessionID)
}

// AddSession records new session information
func (d *DB) AddSession(uuid string, sessionID string, IPAddr string, DeviceData string) error {
	dateint := time.Now().Unix()
	expdateint := time.Now().Add(time.Hour * loginSessionExpHours).Unix()
	_, err := d.db.Exec(`INSERT INTO session VALUES($1,$2,$3,$4,$5,$6,$7)`, uuid, sessionID, dateint, dateint, expdateint, IPAddr, DeviceData)
	if err != nil {
		return err
	}
	return nil
}

// GetSession deletes the session
func (d *DB) GetSession(uuid string) ([]Session, error) {
	res := []Session{}
	rows, err := d.db.Query("SELECT * FROM session WHERE uuid = $1", uuid)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		s := Session{}
		err = rows.Scan(&s.UUID, &s.SessionID, &s.LoginDate,
			&s.LastDate, &s.ExpireDate, &s.IPAddr, &s.DeviceData)
		if err != nil {
			return []Session{}, err
		}
		res = append(res, s)
	}
	return res, nil
}

// RemoveSession deletes the session
func (d *DB) RemoveSession(uuid string, sessionID string) error {
	_, err := d.db.Exec(`DELETE FROM session WHERE uuid = $1 AND sessionid = $2`, uuid, sessionID)
	if err != nil {
		return err
	}
	return nil
}

// UpdateSessionLastUsed deletes the session
func (d *DB) UpdateSessionLastUsed(uuid string, sessionID string) error {
	dateint := time.Now().Unix()
	_, err := d.db.Exec(`UPDATE session SET lastdate = $1 WHERE uuid = $2 AND sessionid = $3`, dateint, uuid, sessionID)
	if err != nil {
		return err
	}
	return nil
}

// VerifyToken verifies JWT and returns UUID and sessionID of JWT holder
func (d *DB) VerifyToken(token string) (string, string, error) {
	var claims jwt.StandardClaims

	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		return verifyKey, nil
	})
	if err != nil {
		return "", "", ErrInvalidToken
	}

	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return "", "", ErrInvalidToken
	}
	var expdate int64
	r := d.db.QueryRow("SELECT expiredate FROM session WHERE uuid = $1 AND sessionid = $2", claims.Audience, claims.Id)
	err = r.Scan(&expdate)
	if err == sql.ErrNoRows {
		return "", "", ErrInvalidToken
	} else if err != nil {
		return "", "", err
	}
	if time.Now().Unix() > expdate {
		return "", "", ErrInvalidToken
	}
	return claims.Audience, claims.Id, nil
}

// GenerateInviteToken generates token for invitation
func (d *DB) GenerateInviteToken(useruuid string) (string, error) {
	expdateint := time.Now().Add(time.Hour * verifyTokenExpHours).Unix()
	newtoken, err := GenerateID(IDTypeVerifyToken)
	if err != nil {
		return "", err
	}
	_, err = d.db.Exec(`INSERT INTO invitetoken VALUES($1,$2,$3)`, useruuid, newtoken, expdateint)
	if err != nil {
		return "", err
	}
	return newtoken, nil
}

// CheckInviteToken checks invitation token
func (d *DB) CheckInviteToken(token string) error {
	cnt := 0
	dateint := time.Now().Unix()
	r := d.db.QueryRow("SELECT count(*) FROM invitetoken WHERE token = $1 AND expdate > $2", token, dateint)
	err := r.Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if cnt != 1 {
		return ErrInvalidToken
	}
	return nil
}

// DeleteInviteToken removes invitation token
func (d *DB) DeleteInviteToken(token string) error {
	_, err := d.db.Exec(`DELETE FROM invitetoken WHERE token = $1`, token)
	if err != nil {
		return err
	}
	return nil
}

// PreRegistUser records new user information into preuser and return token to activate
func (d *DB) PreRegistUser(username string, email string, password string) (string, error) {
	dateint := time.Now().Unix()

	uuid := ""
	//Check user email
	r := d.db.QueryRow("SELECT uuid FROM preuser WHERE email = $1 AND expdate > $2 UNION SELECT uuid FROM auth WHERE email = $1", email, dateint)
	err := r.Scan(&uuid)
	if err == nil {
		return "", ErrExistUser
	} else if err != sql.ErrNoRows {
		return "", err
	}
	//Check user username
	r = d.db.QueryRow("SELECT uuid FROM preuser WHERE username = $1 AND expdate > $2 UNION SELECT uuid FROM username WHERE username = $1", username, dateint)
	err = r.Scan(&uuid)
	if err == nil {
		return "", ErrExistUser
	} else if err != sql.ErrNoRows {
		return "", err
	}

	expdateint := time.Now().Add(time.Hour * verifyTokenExpHours).Unix()

	newuuid, err := GenerateID(IDTypeUser)
	if err != nil {
		return "", err
	}

	newsalt, err := GenerateID(IDTypeSalt)
	if err != nil {
		return "", err
	}

	newtoken, err := GenerateID(IDTypeVerifyToken)
	if err != nil {
		return "", err
	}

	r = d.db.QueryRow("select uuid from preuser where uuid = $1 and expdate > $2 union select uuid from auth where uuid=$1", newuuid, dateint)
	err = r.Scan(&uuid)
	if err == nil {
		return "", errors.New("Duplicate UUID is detected. You're so unlucky")
	} else if err != sql.ErrNoRows {
		return "", err
	}
	_, err = d.db.Exec(`INSERT INTO preuser VALUES($1,$2,$3,$4,$5,$6,$7)`, newuuid, username, email, passhash(password, newsalt), newsalt, newtoken, expdateint)
	if err != nil {
		return "", err
	}

	return newtoken, nil
}

// RegistUser activates preuser
func (d *DB) RegistUser(token string) error {
	var uuid string
	var username string
	var email string
	var password string
	var salt string

	dateint := time.Now().Unix()
	r := d.db.QueryRow("SELECT uuid,username,email,password,salt FROM preuser WHERE token = $1 AND expdate > $2", token, dateint)
	err := r.Scan(&uuid, &username, &email, &password, &salt)
	if err == sql.ErrNoRows {
		return ErrInvalidToken
	} else if err != nil {
		return err
	}

	fid, err := GenerateID(IDTypeFolder)
	if err != nil {
		return err
	}
	userfid, err := d.GetUserFID()
	if err != nil {
		return err
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM preuser WHERE token = $1`, token)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	// Add user login information
	_, err = tx.Exec(`INSERT INTO username VALUES($1,$2)`, uuid, username)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	_, err = tx.Exec(`INSERT INTO auth VALUES($1,$2,$3,$4)`, uuid, email, password, salt)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	// Add user profile
	_, err = tx.Exec(`INSERT INTO profile VALUES($1,'','',$2,'','ja')`, uuid, dateint)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	// Add user folder
	_, err = tx.Exec(`INSERT INTO folder VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
		fid, uuid, userfid, username, FilePermPrivate, dateint, dateint, uuid)
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

// ChangePass checks oldpass and changes to new pass
func (d *DB) ChangePass(uuid string, oldpass string, newpass string) error {
	dbpass := ""
	salt := ""

	r := d.db.QueryRow("SELECT password,salt FROM auth WHERE uuid = $1", uuid)
	err := r.Scan(&dbpass, &salt)
	if err == sql.ErrNoRows {
		return ErrIDPassInvalid
	} else if err != nil {
		return err
	}

	if passhash(oldpass, salt) != dbpass {
		return ErrIDPassInvalid
	}

	err = d.SetPass(uuid, newpass)
	if err != nil {
		return err
	}
	return nil
}

// SetPass changes to new pass
func (d *DB) SetPass(uuid string, newpass string) error {
	newsalt, err := GenerateID(IDTypeSalt)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(`UPDATE auth SET password = $1, salt = $2 WHERE uuid = $3`, passhash(newpass, newsalt), newsalt, uuid)
	if err != nil {
		return err
	}
	return nil
}

// ResetPass generates and returns token to reset password
func (d *DB) ResetPass(email string) (string, string, error) {
	uuid := ""

	r := d.db.QueryRow("SELECT uuid FROM auth WHERE email = $1", email)
	err := r.Scan(&uuid)
	if err == sql.ErrNoRows {
		return "", "", nil
	} else if err != nil {
		return "", "", err
	}

	expdateint := time.Now().Add(time.Hour * verifyTokenExpHours).Unix()
	token, err := GenerateID(IDTypeVerifyToken)
	if err != nil {
		return "", "", err
	}

	_, err = d.db.Exec(`INSERT INTO passreset VALUES($1,$2,$3)`, uuid, token, expdateint)
	if err != nil {
		return "", "", err
	}
	return uuid, token, nil
}

// ResetPassTokenCheck checks the token to reset password and returns UUID
func (d *DB) ResetPassTokenCheck(token string) (string, error) {
	dateint := time.Now().Unix()
	uuid := ""

	r := d.db.QueryRow("SELECT uuid FROM passreset WHERE token = $1 AND expdate > $2", token, dateint)
	err := r.Scan(&uuid)
	if err == sql.ErrNoRows {
		return "", ErrInvalidToken
	} else if err != nil {
		return "", err
	}
	return uuid, nil
}

// ResetPassVerify checks the token and changes to new pass
func (d *DB) ResetPassVerify(token string, newpass string) (string, error) {
	uuid, err := d.ResetPassTokenCheck(token)
	if err != nil {
		return "", err
	}

	err = d.SetPass(uuid, newpass)
	if err != nil {
		return "", err
	}

	_, err = d.db.Exec(`DELETE FROM passreset WHERE token = $1`, token)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

// IsUserLocked checks the user is locked.
func (d *DB) IsUserLocked(uuid string) (bool, error) {
	pass := ""
	r := d.db.QueryRow("SELECT password FROM auth WHERE uuid = $1", uuid)
	err := r.Scan(&pass)
	if err == sql.ErrNoRows {
		return true, ErrIDPassInvalid
	} else if err != nil {
		return true, err
	}

	if pass == "" {
		return true, nil
	}
	return false, nil
}

// GetLogs returns the list of logs
// target is list of teamids
func (d *DB) GetLogs(offset int, limit int, uuid string, target []string, ltype []string) ([]Log, error) {
	if len(target) == 0 {
		return []Log{}, nil
	}
	params := []interface{}{}
	// TargetUUID in target
	params = append(params, uuid)
	sql := "SELECT * FROM log WHERE (targetuuid IN ($" + strconv.Itoa(len(params))
	for _, v := range target {
		params = append(params, v)
		sql += ",$" + strconv.Itoa(len(params))
	}
	// UUID in target
	params = append(params, uuid)
	sql += ") OR uuid IN ($" + strconv.Itoa(len(params)) + "))"
	// Log type in ltype
	if len(ltype) > 0 {
		sql += " AND type IN ("
		for _, v := range ltype {
			params = append(params, v)
			sql += "$" + strconv.Itoa(len(params)) + ","
		}
		sql = strings.TrimRight(sql, ",")
		sql += ")"
	}
	// Sort
	sql += " ORDER BY date DESC,uuid,type"
	// Limit and offset

	params = append(params, limit)
	sql += " LIMIT $" + strconv.Itoa(len(params))
	if offset > 0 {
		params = append(params, offset)
		sql += " OFFSET $" + strconv.Itoa(len(params))
	}

	var res []Log
	rows, err := d.db.Query(sql, params...)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		var l Log
		err = rows.Scan(&l.UUID, &l.Date, &l.Type, &l.IPAddr, &l.SessionID,
			&l.TargetUUID, &l.TargetFDID, &l.ExtDataID)
		if err != nil {
			return res, err
		}
		l.Date /= 1000000000
		res = append(res, l)
	}
	if err = rows.Err(); err != nil {
		return res, err
	}
	return res, nil
}

// GetLoginPassResetLog returns log of loginpassreset
func (d *DB) GetLoginPassResetLog(logid int64) (LogExtLoginPassReset, error) {
	var res LogExtLoginPassReset
	row := d.db.QueryRow("SELECT * FROM logextloginpassreset WHERE id = $1", logid)
	err := row.Scan(&res.ID, &res.DeviceData)
	if err != nil {
		return res, err
	}
	return res, nil
}

// AddLogLogin adds login log
func (d *DB) AddLogLogin(uuid string, sessionID string, ipaddr string, devinfo string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	var extdataid int64
	err = tx.QueryRow(`INSERT INTO logextloginpassreset (devicedata) VALUES ($1) RETURNING id`, devinfo).Scan(&extdataid)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	_, err = tx.Exec(`INSERT INTO log VALUES ($1,$2,$3,$4,$5,'','',$6)`,
		uuid, time.Now().UnixNano(), LogTypeAuthLogin, ipaddr, sessionID, extdataid)
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

// AddLogPassReset adds pass reset log
func (d *DB) AddLogPassReset(uuid string, ipaddr string, devinfo string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	var extdataid int64
	err = tx.QueryRow(`INSERT INTO logextloginpassreset (devicedata) VALUES ($1) RETURNING id`, devinfo).Scan(&extdataid)
	if err != nil {
		if re := tx.Rollback(); re != nil {
			err = fmt.Errorf("%s: %w", re.Error(), err)
		}
		return err
	}
	_, err = tx.Exec(`INSERT INTO log VALUES ($1,$2,$3,$4,'','','',$5)`,
		uuid, time.Now().UnixNano(), LogTypeAuthPassReset, ipaddr, extdataid)
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

// AddLogPassChange adds pass change log
func (d *DB) AddLogPassChange(uuid string, ipaddr string, sessionID string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO log VALUES ($1,$2,$3,$4,$5,'','',-1)`,
		uuid, time.Now().UnixNano(), LogTypeAuthPassChange, ipaddr, sessionID)
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

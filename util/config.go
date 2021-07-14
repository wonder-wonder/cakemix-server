package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Default
var (
	dbHost                    = ""
	dbPort                    = ""
	dbUser                    = ""
	dbPass                    = ""
	dbName                    = ""
	apiHost                   = "localhost"
	apiPort                   = "8081"
	apiCORS                   = ""
	apiPermitUserToCreateTeam = false
	frontDir                  = "/usr/share/cakemix/www"
	dataDir                   = "/var/lib/cakemix/cmdat"
	signPubKey                = "/etc/cakemix/keys/signkey.pub"
	signPrvKey                = "/etc/cakemix/keys/signkey"
	logFile                   = ""
	sendgridAPIKey            = ""
	fromAddr                  = "cakemix@localhost"
	fromName                  = "Cakemix"
	tmplResetPW               = "/usr/share/cakemix/mail/resetpw.tmpl"
	tmplRegist                = "/usr/share/cakemix/mail/regist.tmpl"
)

// DBConf is structure for database configuration
type DBConf struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

// APIConf is structure for API configuration
type APIConf struct {
	Host                   string
	Port                   string
	CORS                   string
	PermitUserToCreateTeam bool
}

// FileConf is structure for file configuration
type FileConf struct {
	FrontDir   string
	DataDir    string
	SignPubKey string
	SignPrvKey string
	LogFile    string
}

// MailConf is structure for mail configuration
type MailConf struct {
	SendGridAPIKey string
	FromAddr       string
	FromName       string
	TmplResetPW    string
	TmplRegist     string
}

// LoadConfigEnv reads config from environment variable
func LoadConfigEnv() {
	// DB config
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

	// Mail config
	if os.Getenv("SENDGRID_API_KEY") != "" {
		sendgridAPIKey = os.Getenv("SENDGRID_API_KEY")
	}
}

// LoadConfigFile reads config from file
func LoadConfigFile(path string) error {
	// #nosec G304
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(raw), "\n")
	for _, v := range lines {
		// Remove comment
		line := strings.SplitN(v, "#", 2)[0]
		// Remove leading and trailing extra spaces
		line = strings.Trim(line, " ")
		// Ignore empty line
		if len(line) == 0 {
			continue
		}
		confs := strings.SplitN(v, " ", 2)
		if len(confs) != 2 {
			confs = append(confs, "")
		}
		confkey := strings.Trim(confs[0], " ")
		confvalue := strings.Trim(confs[1], " ")
		switch strings.ToLower(confkey) {
		case "dbhost":
			dbHost = confvalue
		case "dbport":
			dbPort = confvalue
		case "dbuser":
			dbUser = confvalue
		case "dbpass":
			dbPass = confvalue
		case "dbname":
			dbName = confvalue
		case "apihost":
			apiHost = confvalue
		case "apiport":
			apiPort = confvalue
		case "apicors":
			apiCORS = confvalue
		case "permitusertocreateteam":
			confstrlower := strings.ToLower(confvalue)
			if confstrlower == "no" || confstrlower == "false" || confstrlower == "disable" {
				apiPermitUserToCreateTeam = false
			}
			if confstrlower == "yes" || confstrlower == "true" || confstrlower == "enable" {
				apiPermitUserToCreateTeam = true
			}
		case "frontdir":
			frontDir = confvalue
		case "datadir":
			dataDir = confvalue
		case "signpubkey":
			signPubKey = confvalue
		case "signprvkey":
			signPrvKey = confvalue
		case "logfile":
			logFile = confvalue
		case "mailsgapikey":
			sendgridAPIKey = confvalue
		case "mailfromaddr":
			fromAddr = confvalue
		case "mailfromname":
			fromName = confvalue
		case "mailresetpwtmpl":
			tmplResetPW = confvalue
		case "mailregisttmpl":
			tmplRegist = confvalue
		default:
			return fmt.Errorf("unknown option: %v", confkey)
		}
	}
	return nil
}

// GetDBConf returns database config
func GetDBConf() DBConf {
	return DBConf{
		Host: dbHost,
		Port: dbPort,
		User: dbUser,
		Pass: dbPass,
		Name: dbName,
	}
}

// GetAPIConf returns API config
func GetAPIConf() APIConf {
	return APIConf{
		Host:                   apiHost,
		Port:                   apiPort,
		CORS:                   apiCORS,
		PermitUserToCreateTeam: apiPermitUserToCreateTeam,
	}
}

// GetFileConf returns file config
func GetFileConf() FileConf {
	return FileConf{
		FrontDir:   frontDir,
		DataDir:    dataDir,
		SignPubKey: signPubKey,
		SignPrvKey: signPrvKey,
		LogFile:    logFile,
	}
}

// GetMailConf returns mail config
func GetMailConf() MailConf {
	return MailConf{
		SendGridAPIKey: sendgridAPIKey,
		FromAddr:       fromAddr,
		FromName:       fromName,
		TmplResetPW:    tmplResetPW,
		TmplRegist:     tmplRegist,
	}
}

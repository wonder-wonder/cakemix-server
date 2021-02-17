package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Default
var (
	dbHost         = ""
	dbPort         = ""
	dbUser         = ""
	dbPass         = ""
	dbName         = ""
	apiHost        = ""
	apiPort        = "8081"
	frontDir       = "./"
	dataDir        = "./cmdat"
	signPubKey     = "./signkey.pub"
	signPrvKey     = "./signkey"
	sendgridAPIKey = ""
	fromAddr       = "cakemix@wonder-wonder.xyz"
	fromName       = "Cakemix"
)

type DBConf struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}
type APIConf struct {
	Host string
	Port string
}
type FileConf struct {
	FrontDir   string
	DataDir    string
	SignPubKey string
	SignPrvKey string
}
type MailConf struct {
	SendGridAPIKey string
	FromAddr       string
	FromName       string
}

func LoadConfig() {
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

	// API config
	if os.Getenv("APIADDR") != "" {
		apiHost = os.Getenv("APIADDR")
	}
	if os.Getenv("PORT") != "" {
		apiPort = os.Getenv("PORT")
	}

	// File config
	if os.Getenv("FRONTDIR") != "" {
		frontDir = os.Getenv("FRONTDIR")
	}
	if os.Getenv("DATADIR") != "" {
		dataDir = os.Getenv("DATADIR")
	}
	if os.Getenv("SIGNPRVKEY") != "" {
		signPrvKey = os.Getenv("SIGNPRVKEY")
	}
	if os.Getenv("SIGNPUBKEY") != "" {
		signPubKey = os.Getenv("SIGNPUBKEY")
	}

	// Mail config
	sendgridAPIKey = ""
	if os.Getenv("SENDGRID_API_KEY") != "" {
		sendgridAPIKey = os.Getenv("SENDGRID_API_KEY")
	}
	fromAddr = "cakemix@wonder-wonder.xyz"
	fromName = "Cakemix"
}

func LoadConfigFile(path string) error {
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
			return errors.New("Value is not specified")
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
		case "frontdir":
			frontDir = confvalue
		case "datadir":
			dataDir = confvalue
		case "signpubkey":
			signPubKey = confvalue
		case "signprvkey":
			signPrvKey = confvalue
		case "mailsgapikey":
			sendgridAPIKey = confvalue
		case "mailfromaddr":
			fromAddr = confvalue
		case "mailfromname":
			fromName = confvalue
		default:
			return fmt.Errorf("Unknown option: %v", confkey)
		}
	}
	return nil
}

func GetDBConf() DBConf {
	return DBConf{
		Host: dbHost,
		Port: dbPort,
		User: dbUser,
		Pass: dbPass,
		Name: dbName,
	}
}
func GetAPIConf() APIConf {
	return APIConf{
		Host: apiHost,
		Port: apiPort,
	}
}
func GetFileConf() FileConf {
	return FileConf{
		FrontDir:   frontDir,
		DataDir:    dataDir,
		SignPubKey: signPubKey,
		SignPrvKey: signPrvKey,
	}
}
func GetMailConf() MailConf {
	return MailConf{
		SendGridAPIKey: sendgridAPIKey,
		FromAddr:       fromAddr,
		FromName:       fromName,
	}
}

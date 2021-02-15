package util

import "os"

// Default
var (
	dbHost         = ""
	dbPort         = ""
	dbUser         = ""
	dbPass         = ""
	dbName         = ""
	apiHost        = "localhost"
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

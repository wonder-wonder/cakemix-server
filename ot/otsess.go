package ot

import (
	"github.com/gorilla/websocket"
	"github.com/wonder-wonder/cakemix-server/db"
)

// Session is structure for OT session
type Session struct {
	db                 *db.DB
	incnum             int
	saveRequest        chan bool
	isSaveTimerRunning bool
	lastUpdater        string
	lastGCRev          int
	panicStop          chan bool

	DocID   string
	DocInfo db.Document
	Clients map[string]*Client
	request chan Request
	OT      *OT
}

// Client is structure for client connection
type Client struct {
	conn     *websocket.Conn
	sess     *Session
	lastRev  int
	readOnly bool

	response chan Response
	ClientID string
	UserInfo struct {
		UUID string
		Name string
	}
	Selection []SelData
}

// Request is structure for OT session request
type Request struct {
	Type     WSMsgType
	ClientID string
	Data     interface{}
}

// Response is structure for OT session response
type Response struct {
	Type WSMsgType
	Data interface{}
}

package ot

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"unicode/utf16"

	"github.com/gorilla/websocket"
	"github.com/wonder-wonder/cakemix-server/db"
)

const (
	autoSaveInterval  = 60  //Sec
	otHistGCThreshold = 200 //Ops
)

// OT session management

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

// NewOTClient generates OT client data
func NewOTClient(conn *websocket.Conn, uuid string, name string, readOnly bool) *Client {
	otc := Client{}
	otc.conn = conn
	otc.readOnly = readOnly

	otc.response = make(chan Response)
	otc.UserInfo.UUID = uuid
	otc.UserInfo.Name = name
	otc.Selection = []SelData{}
	return &otc
}

// Close finishes OT client data
func (cl *Client) Close() {
	close(cl.response)
}

// Session is structure for OT session
type Session struct {
	db                 *db.DB
	incnum             int
	saveRequest        chan bool
	isSaveTimerRunning bool
	lastUpdater        string
	lastGCRev          int

	DocID   string
	DocInfo db.Document
	Clients map[string]*Client
	request chan Request
	OT      *OT
}

var (
	otSessions     = map[string]*Session{}
	otSessionsLock = make(chan bool, 1)
)

// OpenSession returns OT session. If unavailable, generates new session.
func OpenSession(db *db.DB, docID string) (*Session, error) {
	otSessionsLock <- true
	defer func() { <-otSessionsLock }()
	if v, ok := otSessions[docID]; ok {
		return v, nil
	}
	ots := &Session{}
	otSessions[docID] = ots
	ots.db = db
	ots.incnum = 0
	ots.saveRequest = make(chan bool)
	ots.lastGCRev = 0

	ots.DocID = docID
	docInfo, err := db.GetDocumentInfo(docID)
	if err != nil {
		return nil, err
	}
	ots.DocInfo = docInfo
	ots.Clients = map[string]*Client{}
	ots.request = make(chan Request)

	text, err := db.GetLatestDocument(docID)
	if err != nil {
		return nil, err
	}
	ots.OT = NewOT(text)

	go ots.SessionLoop()

	return ots, nil
}

// Close finishes OT session
func (sess *Session) Close() {
	sess.isSaveTimerRunning = false
	close(sess.saveRequest)
	otSessionsLock <- true
	defer func() { <-otSessionsLock }()
	delete(otSessions, sess.DocID)
	if len(sess.OT.History) > 0 {
		updateruuid := sess.lastUpdater
		err := sess.db.SaveDocument(sess.DocID, updateruuid, sess.OT.Text)
		if err != nil {
			log.Printf("OT session close error: %v\n", err)
		}
		err = sess.db.UpdateDocument(sess.DocID, updateruuid)
		if err != nil {
			log.Printf("OT session close error: %v\n", err)
		}
		fmt.Printf("Session(%s) closed (total %d ops): ", sess.DocID, sess.OT.Revision)
		if len(sess.OT.Text) <= 10 {
			fmt.Printf("%s\n", sess.OT.Text)
		} else {
			fmt.Printf("%s...\n", sess.OT.Text[:9])
		}
	}
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

// SessionLoop is main loop for OT session
func (sess *Session) SessionLoop() {
	for {
		select {
		case req := <-sess.request:
			if req.Type == OTReqResTypeJoin {
				if v, ok := req.Data.(*Client); ok {
					v.sess = sess
					sess.incnum++
					v.ClientID = strconv.Itoa(sess.incnum)
					sess.Clients[v.ClientID] = v
					v.Response(WSMsgTypeOK, nil)
				}
			} else if req.Type == WSMsgTypeDoc {
				res := DocData{Clients: map[string]ClientData{}, Document: sess.OT.Text, Revision: sess.OT.Revision}
				for _, cl := range sess.Clients {
					if cl.ClientID == req.ClientID {
						continue
					}
					rescl := ClientData{Name: cl.UserInfo.Name}
					rescl.Selection.Ranges = []SelData{}
					for _, sel := range cl.Selection {
						rescl.Selection.Ranges = append(rescl.Selection.Ranges, sel)
					}
					res.Clients[cl.ClientID] = rescl
				}
				sess.Clients[req.ClientID].lastRev = res.Revision
				res.Owner = sess.DocInfo.OwnerUUID
				res.Permission = int(sess.DocInfo.Permission)
				res.Editable = !sess.Clients[req.ClientID].readOnly
				go sess.Clients[req.ClientID].Response(WSMsgTypeDoc, res)
			} else if req.Type == WSMsgTypeOp {
				opdat, ok := req.Data.(OpData)
				if !ok {
					continue
				}
				ops := Ops{User: req.ClientID, Ops: []Op{}}
				for _, op := range opdat.Operation {
					switch opt := op.(type) {
					case float64:
						opi := int(opt)
						if opi < 0 {
							ops.Ops = append(ops.Ops, Op{OpType: OpTypeDelete, Len: -opi})
						} else {
							ops.Ops = append(ops.Ops, Op{OpType: OpTypeRetain, Len: opi})
						}
					case string:
						ops.Ops = append(ops.Ops, Op{OpType: OpTypeInsert, Len: len(utf16.Encode([]rune(opt))), Text: opt})
					default:
						continue
					}
				}
				optrans, err := sess.OT.Operate(opdat.Revision, ops)
				if err != nil {
					log.Printf("OT session error: operate error: %v\n", err)
				}
				opraw := []interface{}{}
				for _, v := range optrans.Ops {
					if v.OpType == OpTypeRetain {
						opraw = append(opraw, v.Len)
					} else if v.OpType == OpTypeInsert {
						opraw = append(opraw, v.Text)
					} else if v.OpType == OpTypeDelete {
						opraw = append(opraw, -v.Len)
					}
				}
				opdat.Operation = opraw

				sess.Clients[req.ClientID].Selection = opdat.Selection.Ranges
				sess.Clients[req.ClientID].lastRev = sess.OT.Revision

				opres := []interface{}{req.ClientID, opdat.Operation, opdat.Selection}
				for cid, v := range sess.Clients {
					if cid == req.ClientID {
						go v.Response(WSMsgTypeOK, nil)
						continue
					}
					go v.Response(WSMsgTypeOp, opres)
				}

				sess.lastUpdater = sess.Clients[req.ClientID].UserInfo.UUID
				sess.lastGCRev++

				// Start autosave timer
				go func() {
					defer func() {
						if r := recover(); r != nil {
							return
						}
					}()

					if sess.isSaveTimerRunning {
						return
					}
					sess.isSaveTimerRunning = true
					time.Sleep(time.Second * autoSaveInterval)
					if !sess.isSaveTimerRunning {
						return
					}
					sess.isSaveTimerRunning = false
					sess.saveRequest <- true
				}()
				if sess.lastGCRev >= otHistGCThreshold {
					min := sess.OT.Revision
					for _, c := range sess.Clients {
						if c.lastRev < min {
							min = c.lastRev
						}
					}
					for i := sess.OT.Revision - len(sess.OT.History); i < min-1; i++ {
						delete(sess.OT.History, i)
					}
					sess.lastGCRev = 0
					fmt.Printf("Session(%s) OT GC: rev is %d, hist len is %d\n", sess.DocID, sess.OT.Revision, len(sess.OT.History))
				}
			} else if req.Type == WSMsgTypeSel {
				seldat, ok := req.Data.(Ranges)
				if !ok {
					continue
				}
				selresdat := Ranges{}
				selresdat.Ranges = seldat.Ranges
				for cid, v := range sess.Clients {
					if cid == req.ClientID {
						continue
					}
					go v.Response(WSMsgTypeSel, []interface{}{req.ClientID, selresdat})
				}
			} else if req.Type == WSMsgTypeQuit {
				for cid, v := range sess.Clients {
					if cid == req.ClientID {
						delete(sess.Clients, cid)
						continue
					}
					go v.Response(WSMsgTypeQuit, req.ClientID)
				}
				if len(sess.Clients) == 0 {
					sess.Close()
					return
				}
			}
		case <-sess.saveRequest:
			if len(sess.OT.History) > 0 {
				updateruuid := sess.lastUpdater
				err := sess.db.SaveDocument(sess.DocID, updateruuid, sess.OT.Text)
				if err != nil {
					log.Printf("OT session error: save error: %v\n", err)
				}
				err = sess.db.UpdateDocument(sess.DocID, updateruuid)
				if err != nil {
					log.Printf("OT session error: save error: %v\n", err)
				}
				fmt.Printf("Session(%s) auto saved (total %d ops): ", sess.DocID, sess.OT.Revision)
				if len(sess.OT.Text) <= 10 {
					fmt.Printf("%s\n", sess.OT.Text)
				} else {
					fmt.Printf("%s...\n", sess.OT.Text[:9])
				}
			}
		}
	}
}

// ClientLoop is main loop for OT client
func (cl *Client) ClientLoop() {
	request := make(chan []byte)
	// Reader routine
	cancel := false
	defer func() { cancel = true }()
	go func() {
		for !cancel {
			_, msg, err := cl.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
					close(request)
					return
				}
				log.Printf("OT client error: read error: %v\n", err)
				close(request)
				return
			}
			request <- msg
		}
	}()
	for {
		select {
		case req, ok := <-request:
			if !ok {
				cl.sess.Request(WSMsgTypeQuit, cl.ClientID, nil)
				return
			}
			if cl.readOnly {
				log.Printf("OT client error: permission denied\n")
				cl.sess.Request(WSMsgTypeQuit, cl.ClientID, nil)
				return
			}
			mtype, dat, err := parseMsg(req)
			if err != nil {
				log.Printf("OT client error: %v\n", err)
				panic(err)
			}
			if mtype == WSMsgTypeOp {
				opdat, ok := dat.(OpData)
				if !ok {
					panic("Logic error")
				}
				cl.sess.Request(WSMsgTypeOp, cl.ClientID, opdat)
			} else if mtype == WSMsgTypeSel {
				opdat, ok := dat.(Ranges)
				if !ok {
					panic("Logic error")
				}
				cl.sess.Request(WSMsgTypeSel, cl.ClientID, opdat)
			}
		case resdat := <-cl.response:
			resraw, err := convertToMsg(resdat.Type, resdat.Data)
			if err != nil {
				log.Printf("OT client error: response error: %v\n", err)
				panic(err)
			}
			cl.conn.WriteMessage(websocket.TextMessage, resraw)
		}
	}
}

// Request requests to session
func (sess *Session) Request(t WSMsgType, cid string, dat interface{}) {
	sess.request <- Request{Type: t, ClientID: cid, Data: dat}
}

// Response responds to client
func (cl *Client) Response(t WSMsgType, dat interface{}) {
	cl.response <- Response{Type: t, Data: dat}
}

// AddClient requests to server to add new client
func (sess *Session) AddClient(cl *Client) {
	sess.Request(OTReqResTypeJoin, "", cl)
	<-cl.response
}

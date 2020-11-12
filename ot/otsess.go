package ot

import (
	"fmt"
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
// OTClient is structure for client connection
type OTClient struct {
	conn    *websocket.Conn
	sess    *OTSession
	lastRev int

	response chan OTResponse
	ClientID string
	UserInfo struct {
		UUID string
		Name string
	}
	Selection []SelData
}

func NewOTClient(conn *websocket.Conn, uuid string, name string) *OTClient {
	otc := OTClient{}
	otc.conn = conn
	otc.response = make(chan OTResponse)
	otc.UserInfo.UUID = uuid
	otc.UserInfo.Name = name
	otc.Selection = []SelData{}
	return &otc
}

func (cl *OTClient) Close() {
	close(cl.response)
}

// OTClient is structure for OT session
type OTSession struct {
	db                 *db.DB
	incnum             int
	saveRequest        chan bool
	isSaveTimerRunning bool
	lastUpdater        string
	lastGCRev          int

	DocID   string
	DocInfo db.Document
	Clients map[string]*OTClient
	request chan OTRequest
	OT      *OT
}

var (
	otSessions     = map[string]*OTSession{}
	otSessionsLock = make(chan bool, 1)
)

func OpenOTSession(db *db.DB, docID string) (*OTSession, error) {
	otSessionsLock <- true
	defer func() { <-otSessionsLock }()
	if v, ok := otSessions[docID]; ok {
		return v, nil
	}
	ots := &OTSession{}
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
	ots.Clients = map[string]*OTClient{}
	ots.request = make(chan OTRequest)

	// TODO: restore OT
	text, err := db.GetLatestDocument(docID)
	if err != nil {
		return nil, err
	}
	ots.OT = NewOT(text)

	go ots.SessionLoop()

	return ots, nil
}

func (sess *OTSession) Close() {
	sess.isSaveTimerRunning = false
	close(sess.saveRequest)
	otSessionsLock <- true
	defer func() { <-otSessionsLock }()
	delete(otSessions, sess.DocID)
	if len(sess.OT.History) > 0 {
		updateruuid := sess.lastUpdater
		err := sess.db.SaveDocument(sess.DocID, updateruuid, sess.OT.Text)
		if err != nil {
			//TODO
			fmt.Printf("%v\n", err)
		}
		err = sess.db.UpdateDocument(sess.DocID, updateruuid)
		if err != nil {
			//TODO
			fmt.Printf("%v\n", err)
		}
		fmt.Printf("Session(%s) closed (total %d ops): ", sess.DocID, sess.OT.Revision)
		if len(sess.OT.Text) <= 10 {
			fmt.Printf("%s\n", sess.OT.Text)
		} else {
			fmt.Printf("%s...\n", sess.OT.Text[:9])
		}
	}
}

type OTRequest struct {
	Type     WSMsgType
	ClientID string
	Data     interface{}
}
type OTResponse struct {
	Type WSMsgType
	Data interface{}
}

func (sess *OTSession) SessionLoop() {
	for {
		select {
		case req := <-sess.request:
			if req.Type == OTReqResTypeJoin {
				if v, ok := req.Data.(*OTClient); ok {
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
					//TODO
					fmt.Printf("%v\n", err)
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
					//TODO
					fmt.Printf("%v\n", err)
				}
				err = sess.db.UpdateDocument(sess.DocID, updateruuid)
				if err != nil {
					//TODO
					fmt.Printf("%v\n", err)
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

func (cl *OTClient) ClientLoop() {
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
				// TODO
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
				// TODO: close request
				// panic("cclosed")
				cl.sess.Request(WSMsgTypeQuit, cl.ClientID, nil)
				return
			}
			mtype, dat, err := parseMsg(req)
			if err != nil {
				//TODO
				fmt.Printf("%v:%v\n", dat)
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
				//TODO
				panic(err)
			}
			cl.conn.WriteMessage(websocket.TextMessage, resraw)
		}
	}
}
func (cl *OTSession) Request(t WSMsgType, cid string, dat interface{}) {
	cl.request <- OTRequest{Type: t, ClientID: cid, Data: dat}
}
func (cl *OTClient) Response(t WSMsgType, dat interface{}) {
	cl.response <- OTResponse{Type: t, Data: dat}
}
func (sess *OTSession) AddClient(cl *OTClient) {
	sess.Request(OTReqResTypeJoin, "", cl)
	<-cl.response
}

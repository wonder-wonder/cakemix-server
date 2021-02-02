package ot

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"unicode/utf16"

	"github.com/wonder-wonder/cakemix-server/db"
)

const (
	autoSaveInterval     = 60  //Sec
	otHistGCThreshold    = 200 //Ops
	otClientPingInterval = 30  //Sec
)

// OT session management

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
	ots.panicStop = make(chan bool)

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
				showOps(sess.DocID, "req", opdat.Revision, ops)
				optrans, err := sess.OT.Operate(opdat.Revision, ops)
				showOps(sess.DocID, "trans", sess.OT.Revision, optrans)
				if err != nil {
					log.Printf("OT session error: operate error: %v\n", err)
					cl, ok := sess.Clients[req.ClientID]
					if ok {
						cl.Close()
					}
					continue
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
					go func() { sess.panicStop <- true }()
					continue
				}
				err = sess.db.UpdateDocument(sess.DocID, updateruuid)
				if err != nil {
					log.Printf("OT session error: save error: %v\n", err)
					go func() { sess.panicStop <- true }()
					continue
				}
				fmt.Printf("Session(%s) auto saved (total %d ops): ", sess.DocID, sess.OT.Revision)
				if len(sess.OT.Text) <= 10 {
					fmt.Printf("%s\n", sess.OT.Text)
				} else {
					fmt.Printf("%s...\n", sess.OT.Text[:9])
				}
			}
		case <-sess.panicStop:
			for _, v := range sess.Clients {
				v.Close()
			}
			sess.Close()
		}
	}
}

// Request requests to session
func (sess *Session) Request(t WSMsgType, cid string, dat interface{}) {
	sess.request <- Request{Type: t, ClientID: cid, Data: dat}
}

// AddClient requests to server to add new client
func (sess *Session) AddClient(cl *Client) {
	sess.Request(OTReqResTypeJoin, "", cl)
	<-cl.response
}

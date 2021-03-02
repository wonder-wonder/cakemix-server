package ot

import (
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

type SessionServer struct {
	// DB conn
	db *db.DB
	// DocInfo
	docID   string
	docInfo db.Document
	// OT
	ot              *OT
	lastUpdater     string
	countFromLastGC int
	needSave        bool

	// Clients
	clients map[string]*SessionClient

	// Management info
	accumulationClients int // for serial number

	// Channel
	sv2mgr chan SessionServerRequest
	mgr2sv chan SessionManagerRequest
	cl2sv  chan SessionC2SMessage
}

type SessionS2CMessageType int

const (
	SessionS2CMessageTypePing SessionS2CMessageType = iota
	SessionS2CMessageTypeWSMsg
)

type SessionC2SMessageType int

const (
	SessionC2SMessageTypeClose SessionC2SMessageType = iota
	SessionC2SMessageTypeWSMsg
)

type SessionS2CMessage struct {
	msgType SessionS2CMessageType
	message interface{}
}
type SessionC2SMessage struct {
	clientID string
	msgType  SessionC2SMessageType
	message  interface{}
}
type SessionWSMessage struct {
	Event WSMsgType
	Data  interface{}
}

func NewSessionServer(docID string, sv2mgr chan SessionServerRequest, db *db.DB) (*SessionServer, error) {
	sv := &SessionServer{
		db:                  db,
		docID:               docID,
		countFromLastGC:     0,
		needSave:            false,
		clients:             map[string]*SessionClient{},
		accumulationClients: 0,
		sv2mgr:              sv2mgr,
		mgr2sv:              make(chan SessionManagerRequest),
		cl2sv:               make(chan SessionC2SMessage),
	}

	docInfo, err := db.GetDocumentInfo(docID)
	if err != nil {
		return nil, err
	}
	sv.docInfo = docInfo
	text, err := db.GetLatestDocument(docID)
	if err != nil {
		return nil, err
	}
	sv.ot = NewOT(text)

	return sv, nil
}

func (sv *SessionServer) AddClient(clreq *SessionClientRequest) {
	go func() {
		sv.mgr2sv <- SessionManagerRequest{
			reqType: SessionManagerRequestTypeAddClient,
			request: clreq,
		}
	}()
}

func (sv *SessionServer) SendS2M(reqType SessionServerRequestType, request interface{}) {
	go func() {
		sv.sv2mgr <- SessionServerRequest{
			docID:   sv.docID,
			reqType: reqType,
			request: request,
		}
	}()
}

func (sv *SessionServer) Loop() {
	autoSaveTicker := time.NewTicker(time.Second * autoSaveInterval)
	defer autoSaveTicker.Stop()
	sv.SendS2M(SessionServerRequestTypeStarted, nil)
	for {
		select {
		case mgrreq, ok := <-sv.mgr2sv:
			if !ok {
				err := sv.Stop()
				if err != nil {
					log.Printf("OT session close error: %v\n", err)
				}
				return
			}
			switch mgrreq.reqType {
			case SessionManagerRequestTypeAddClient:
				// Add to client list
				clreq, _ := mgrreq.request.(*SessionClientRequest)
				clientID := strconv.Itoa(sv.accumulationClients)
				sv.accumulationClients++
				sv.clients[clientID] = clreq.client

				// Setup client
				clreq.client.clientID = clientID
				clreq.client.cl2sv = sv.cl2sv
				clreq.client.lastRev = sv.ot.Revision

				// Broadcast new client info
				sv.Broadcast(clientID, SessionWSMessage{
					Event: WSMsgTypeJoin,
					Data: ClientJoinData{
						ID:      clreq.client.clientID,
						Name:    clreq.client.profile.Name,
						UUID:    clreq.client.profile.UUID,
						IconURI: clreq.client.profile.IconURI,
					},
				})

				// Finish init and ready
				go func() { clreq.ready <- struct{}{} }()

				// Send doc event to client
				res := DocData{
					Clients:    map[string]ClientData{},
					Document:   sv.ot.Text,
					Revision:   clreq.client.lastRev,
					Owner:      sv.docInfo.OwnerUUID,
					Permission: int(sv.docInfo.Permission),
					Editable:   clreq.client.readOnly,
				}
				for tclientID, cl := range sv.clients {
					if tclientID == clientID {
						continue
					}
					rescl := ClientData{
						Name:    cl.profile.Name,
						UUID:    cl.profile.UUID,
						IconURI: cl.profile.IconURI,
					}
					rescl.Selection.Ranges = []SelData{}
					for _, sel := range cl.selection {
						rescl.Selection.Ranges = append(rescl.Selection.Ranges, sel)
					}
					res.Clients[tclientID] = rescl
				}
				clreq.client.SendS2C(SessionS2CMessageTypeWSMsg, SessionWSMessage{
					Event: WSMsgTypeDoc,
					Data:  res,
				})
			}
		case clreq, _ := <-sv.cl2sv:
			switch clreq.msgType {
			case SessionC2SMessageTypeClose:
				// Closed by client
				sv.CloseClient(clreq.clientID)
				saved, err := sv.SaveDoc()
				if err != nil {
					log.Printf("OT session error: save error: %v\n", err)
					err = sv.Stop()
					if err != nil {
						log.Printf("OT session error: close error: %v\n", err)
					}
					return
				}
				if saved {
					log.Printf("Session(%s) auto saved (total %d ops)", sv.docID, sv.ot.Revision)
				}
			case SessionC2SMessageTypeWSMsg:
				wsmsg := clreq.message.(SessionWSMessage)
				switch wsmsg.Event {
				case WSMsgTypeOp:
					opdat, ok := wsmsg.Data.(OpData)
					if !ok {
						continue
					}
					ops := Ops{User: clreq.clientID, Ops: []Op{}}
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
					optrans, err := sv.ot.Operate(opdat.Revision, ops)
					if err != nil {
						log.Printf("OT session error: operate error: %v\n", err)
						sv.CloseClient(clreq.clientID)
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
					cl := sv.clients[clreq.clientID]

					cl.selection = opdat.Selection.Ranges
					cl.lastRev = sv.ot.Revision

					opres := []interface{}{clreq.clientID, opdat.Operation, opdat.Selection}
					sv.Broadcast(clreq.clientID, SessionWSMessage{
						Event: WSMsgTypeOp,
						Data:  opres,
					})
					cl.SendS2C(SessionS2CMessageTypeWSMsg, SessionWSMessage{
						Event: WSMsgTypeOK,
						Data:  nil,
					})

					sv.lastUpdater = cl.profile.UUID
					sv.countFromLastGC++
					sv.needSave = true

					if sv.countFromLastGC >= otHistGCThreshold {
						sv.countFromLastGC = 0
						min := sv.ot.Revision
						for _, c := range sv.clients {
							if c.lastRev < min {
								min = c.lastRev
							}
						}
						for i := sv.ot.Revision - len(sv.ot.History); i < min-1; i++ {
							delete(sv.ot.History, i)
						}
						log.Printf("Session(%s) OT GC: rev is %d, hist len is %d", sv.docID, sv.ot.Revision, len(sv.ot.History))
					}
				case WSMsgTypeSel:
					seldat, ok := wsmsg.Data.(Ranges)
					if !ok {
						continue
					}
					selresdat := Ranges{}
					selresdat.Ranges = seldat.Ranges
					sv.Broadcast(clreq.clientID, SessionWSMessage{
						Event: WSMsgTypeSel,
						Data:  []interface{}{clreq.clientID, selresdat},
					})
				}
			}
		case <-autoSaveTicker.C:
			saved, err := sv.SaveDoc()
			if err != nil {
				log.Printf("OT session error: save error: %v\n", err)
				err = sv.Stop()
				if err != nil {
					log.Printf("OT session error: close error: %v\n", err)
				}
				return
			}
			if saved {
				log.Printf("Session(%s) auto saved (total %d ops)", sv.docID, sv.ot.Revision)
			}
		}
	}
}

func (sv *SessionServer) Broadcast(from string, message SessionWSMessage) {
	for i, v := range sv.clients {
		if i == from {
			continue
		}
		v.SendS2C(SessionS2CMessageTypeWSMsg, message)
	}
}

func (sv *SessionServer) CloseClient(clientID string) {
	sv.Broadcast(clientID, SessionWSMessage{
		Event: WSMsgTypeQuit,
		Data:  clientID,
	})
	cl := sv.clients[clientID]
	close(cl.sv2cl)
	delete(sv.clients, clientID)
	sv.SendS2M(SessionServerRequestTypeClientClosed, nil)
}

func (sv *SessionServer) SaveDoc() (bool, error) {
	if !sv.needSave {
		return false, nil
	}
	sv.needSave = false
	if len(sv.ot.History) > 0 {
		updateruuid := sv.lastUpdater
		err := sv.db.SaveDocument(sv.docID, updateruuid, sv.ot.Text)
		if err != nil {
			return false, err
		}
		err = sv.db.UpdateDocument(sv.docID, updateruuid)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (sv *SessionServer) Stop() error {
	for i := range sv.clients {
		sv.CloseClient(i)
	}
	_, err := sv.SaveDoc()
	if err != nil {
		return err
	}
	log.Printf("Session(%s) closed (total %d ops)\n", sv.docID, sv.ot.Revision)
	sv.SendS2M(SessionServerRequestTypeStopped, nil)
	return nil
}

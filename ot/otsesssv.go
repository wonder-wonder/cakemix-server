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

// Server is structure for server
type Server struct {
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
	clients map[string]*Client

	// Management info
	accumulationClients int // for serial number

	// Channel
	sv2mgr chan otServerRequest
	mgr2sv chan otManagerRequest
	cl2sv  chan otC2SMessage
}

type otS2CMessageType int

const (
	otS2CMessageTypePing otS2CMessageType = iota
	otS2CMessageTypeWSMsg
)

type otC2SMessageType int

const (
	otC2SMessageTypeClose otC2SMessageType = iota
	otC2SMessageTypeWSMsg
)

type otS2CMessage struct {
	msgType otS2CMessageType
	message interface{}
}
type otC2SMessage struct {
	clientID string
	msgType  otC2SMessageType
	message  interface{}
}
type otWSMessage struct {
	Event WSMsgType
	Data  interface{}
}

// NewServer creates new server
func NewServer(docID string, sv2mgr chan otServerRequest, db *db.DB) (*Server, error) {
	sv := &Server{
		db:                  db,
		docID:               docID,
		countFromLastGC:     0,
		needSave:            false,
		clients:             map[string]*Client{},
		accumulationClients: 0,
		sv2mgr:              sv2mgr,
		mgr2sv:              make(chan otManagerRequest, 10),
		cl2sv:               make(chan otC2SMessage, 100),
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

func (sv *Server) sendS2M(reqType otServerRequestType, request interface{}) {
	go func() {
		sv.sv2mgr <- otServerRequest{
			docID:   sv.docID,
			reqType: reqType,
			request: request,
		}
	}()
}

// Loop is main loop for server
func (sv *Server) Loop() {
	autoSaveTicker := time.NewTicker(time.Second * autoSaveInterval)
	defer autoSaveTicker.Stop()
	sv.sendS2M(otServerRequestTypeStarted, nil)
main:
	for {
		select {
		case mgrreq, ok := <-sv.mgr2sv:
			// Stop request by manager
			if !ok {
				break main
			}
			switch mgrreq.reqType {
			case otManagerRequestTypeAddClient:
				// Add to client list
				clreq, _ := mgrreq.request.(*otClientRequest)
				clientID := strconv.Itoa(sv.accumulationClients)
				sv.accumulationClients++
				sv.clients[clientID] = clreq.client

				// Setup client
				clreq.client.clientID = clientID
				clreq.client.cl2sv = sv.cl2sv
				clreq.client.lastRev = sv.ot.Revision

				// Broadcast new client info
				sv.broadcast(clientID, otWSMessage{
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
				clreq.client.sendS2C(otS2CMessageTypeWSMsg, otWSMessage{
					Event: WSMsgTypeDoc,
					Data:  res,
				})
			}
		case clreq, _ := <-sv.cl2sv:
			switch clreq.msgType {
			case otC2SMessageTypeClose:
				// Closed by client
				sv.closeClient(clreq.clientID)
				saved, err := sv.saveDoc()
				if err != nil {
					log.Printf("OT session error: save error: %v\n", err)
					break main
				}
				if saved {
					log.Printf("Session(%s) auto saved (total %d ops)", sv.docID, sv.ot.Revision)
				}
			case otC2SMessageTypeWSMsg:
				wsmsg := clreq.message.(otWSMessage)
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
						sv.closeClient(clreq.clientID)
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
					sv.broadcast(clreq.clientID, otWSMessage{
						Event: WSMsgTypeOp,
						Data:  opres,
					})
					cl.sendS2C(otS2CMessageTypeWSMsg, otWSMessage{
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
					sv.broadcast(clreq.clientID, otWSMessage{
						Event: WSMsgTypeSel,
						Data:  []interface{}{clreq.clientID, selresdat},
					})
				}
			}
		case <-autoSaveTicker.C:
			saved, err := sv.saveDoc()
			if err != nil {
				log.Printf("OT session error: save error: %v\n", err)
				break main
			}
			if saved {
				log.Printf("Session(%s) auto saved (total %d ops)", sv.docID, sv.ot.Revision)
			}
		}
	}
	for i := range sv.clients {
		sv.closeClient(i)
	}
	_, err := sv.saveDoc()
	if err != nil {
		log.Printf("OT session close error: %v\n", err)
	}
	log.Printf("Session(%s) closed (total %d ops)\n", sv.docID, sv.ot.Revision)
	sv.sendS2M(otServerRequestTypeStopped, nil)
}

func (sv *Server) broadcast(from string, message otWSMessage) {
	for i, v := range sv.clients {
		if i == from {
			continue
		}
		v.sendS2C(otS2CMessageTypeWSMsg, message)
	}
}

func (sv *Server) closeClient(clientID string) {
	sv.broadcast(clientID, otWSMessage{
		Event: WSMsgTypeQuit,
		Data:  clientID,
	})
	cl := sv.clients[clientID]
	close(cl.sv2cl)
	delete(sv.clients, clientID)
	sv.sendS2M(otServerRequestTypeClientClosed, nil)
}

func (sv *Server) saveDoc() (bool, error) {
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

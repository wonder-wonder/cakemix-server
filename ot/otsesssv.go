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

type OTServer struct {
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
	clients map[string]*OTClient

	// Management info
	accumulationClients int // for serial number

	// Channel
	sv2mgr chan OTServerRequest
	mgr2sv chan OTManagerRequest
	cl2sv  chan OTC2SMessage
}

type OTS2CMessageType int

const (
	OTS2CMessageTypePing OTS2CMessageType = iota
	OTS2CMessageTypeWSMsg
)

type OTC2SMessageType int

const (
	OTC2SMessageTypeClose OTC2SMessageType = iota
	OTC2SMessageTypeWSMsg
)

type OTS2CMessage struct {
	msgType OTS2CMessageType
	message interface{}
}
type OTC2SMessage struct {
	clientID string
	msgType  OTC2SMessageType
	message  interface{}
}
type OTWSMessage struct {
	Event WSMsgType
	Data  interface{}
}

func NewOTServer(docID string, sv2mgr chan OTServerRequest, db *db.DB) (*OTServer, error) {
	sv := &OTServer{
		db:                  db,
		docID:               docID,
		countFromLastGC:     0,
		needSave:            false,
		clients:             map[string]*OTClient{},
		accumulationClients: 0,
		sv2mgr:              sv2mgr,
		mgr2sv:              make(chan OTManagerRequest),
		cl2sv:               make(chan OTC2SMessage),
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

func (sv *OTServer) AddClient(clreq *OTClientRequest) {
	go func() {
		sv.mgr2sv <- OTManagerRequest{
			reqType: OTManagerRequestTypeAddClient,
			request: clreq,
		}
	}()
}

func (sv *OTServer) SendS2M(reqType OTServerRequestType, request interface{}) {
	go func() {
		sv.sv2mgr <- OTServerRequest{
			docID:   sv.docID,
			reqType: reqType,
			request: request,
		}
	}()
}

func (sv *OTServer) Loop() {
	autoSaveTicker := time.NewTicker(time.Second * autoSaveInterval)
	defer autoSaveTicker.Stop()
	sv.SendS2M(OTServerRequestTypeStarted, nil)
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
			case OTManagerRequestTypeAddClient:
				// Add to client list
				clreq, _ := mgrreq.request.(*OTClientRequest)
				clientID := strconv.Itoa(sv.accumulationClients)
				sv.accumulationClients++
				sv.clients[clientID] = clreq.client

				// Setup client
				clreq.client.clientID = clientID
				clreq.client.cl2sv = sv.cl2sv
				clreq.client.lastRev = sv.ot.Revision

				// Broadcast new client info
				sv.Broadcast(clientID, OTWSMessage{
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
				clreq.client.SendS2C(OTS2CMessageTypeWSMsg, OTWSMessage{
					Event: WSMsgTypeDoc,
					Data:  res,
				})
			}
		case clreq, _ := <-sv.cl2sv:
			switch clreq.msgType {
			case OTC2SMessageTypeClose:
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
			case OTC2SMessageTypeWSMsg:
				wsmsg := clreq.message.(OTWSMessage)
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
					sv.Broadcast(clreq.clientID, OTWSMessage{
						Event: WSMsgTypeOp,
						Data:  opres,
					})
					cl.SendS2C(OTS2CMessageTypeWSMsg, OTWSMessage{
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
					sv.Broadcast(clreq.clientID, OTWSMessage{
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

func (sv *OTServer) Broadcast(from string, message OTWSMessage) {
	for i, v := range sv.clients {
		if i == from {
			continue
		}
		v.SendS2C(OTS2CMessageTypeWSMsg, message)
	}
}

func (sv *OTServer) CloseClient(clientID string) {
	sv.Broadcast(clientID, OTWSMessage{
		Event: WSMsgTypeQuit,
		Data:  clientID,
	})
	cl := sv.clients[clientID]
	close(cl.sv2cl)
	delete(sv.clients, clientID)
	sv.SendS2M(OTServerRequestTypeClientClosed, nil)
}

func (sv *OTServer) SaveDoc() (bool, error) {
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

func (sv *OTServer) Stop() error {
	for i := range sv.clients {
		sv.CloseClient(i)
	}
	_, err := sv.SaveDoc()
	if err != nil {
		return err
	}
	log.Printf("Session(%s) closed (total %d ops)\n", sv.docID, sv.ot.Revision)
	sv.SendS2M(OTServerRequestTypeStopped, nil)
	return nil
}

package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"unicode/utf16"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/ot"
)

// WSMsgType is WebSocket message type
type WSMsgType int

// WSMsgType list
const (
	WSMsgTypeUnknown WSMsgType = iota
	WSMsgTypeDoc
	WSMsgTypeOp
	WSMsgTypeOK
	WSMsgTypeSel
	WSMsgTypeQuit
	OTReqResTypeJoin
)

// OT Errors
var (
	ErrorInvalidWSMsg = errors.New("WSMessage is invalid")
)

// WSMsg2 is structure for websocket message
type WSMsg2 struct {
	Event string          `json:"e"`
	Data  json.RawMessage `json:"d,omitempty"`
}

type OpData2 struct {
	Revision  int
	Operation []interface{}
	Selection Ranges2
}

// SelData2 is structure for selection data
type SelData2 struct {
	Anchor int `json:"anchor"`
	Head   int `json:"head"`
}
type Ranges2 struct {
	Ranges []SelData2 `json:"ranges"`
}

func parseMsg(rawmsg []byte) (WSMsgType, interface{}, error) {
	msg := WSMsg{}
	err := json.Unmarshal(rawmsg, &msg)
	if err != nil {
		return WSMsgTypeUnknown, nil, ErrorInvalidWSMsg
	}
	if msg.Event == "op" {
		dat := OpData2{}

		// Separate revision, ops, selections
		var rawdat []json.RawMessage
		err := json.Unmarshal(msg.Data, &rawdat)
		if err != nil {
			return WSMsgTypeUnknown, nil, ErrorInvalidWSMsg
		}
		if len(rawdat) != 3 {
			return WSMsgTypeUnknown, nil, ErrorInvalidWSMsg
		}

		// Revision
		revfloat := 0.0
		err = json.Unmarshal(rawdat[0], &revfloat)
		if err != nil {
			return WSMsgTypeUnknown, nil, ErrorInvalidWSMsg
		}
		dat.Revision = int(revfloat)

		// Ops
		err = json.Unmarshal(rawdat[1], &dat.Operation)
		if err != nil {
			return WSMsgTypeUnknown, nil, ErrorInvalidWSMsg
		}

		// Selections
		err = json.Unmarshal(rawdat[2], &dat.Selection)
		if err != nil {
			return WSMsgTypeUnknown, nil, ErrorInvalidWSMsg
		}

		return WSMsgTypeOp, dat, nil
	} else if msg.Event == "sel" {
		dat := Ranges2{}
		// Selections
		err = json.Unmarshal(msg.Data, &dat)
		if err != nil {
			return WSMsgTypeUnknown, nil, ErrorInvalidWSMsg
		}
		return WSMsgTypeSel, dat, nil
		// } else if msg.Event == "ok" {
		// 	return WSMsgTypeOK, nil, nil
		// } else if msg.Event == "doc" {
		// 	return WSMsgTypeDoc, struct{}{}, nil
		// } else if msg.Event == "quit" {
		// 	return WSMsgTypeQuit, string(msg.Data), nil
	}
	return WSMsgTypeUnknown, struct{}{}, nil
}

func convertToMsg(t WSMsgType, dat interface{}) ([]byte, error) {
	var datraw []byte
	if dat != nil {
		var err error
		datraw, err = json.Marshal(dat)
		if err != nil {
			return nil, err
		}
	}
	msg := WSMsg{Data: datraw}
	if t == WSMsgTypeOK {
		msg.Event = "ok"
	} else if t == WSMsgTypeOp {
		msg.Event = "op"
	} else if t == WSMsgTypeDoc {
		msg.Event = "doc"
	} else if t == WSMsgTypeSel {
		msg.Event = "sel"
	} else if t == WSMsgTypeQuit {
		msg.Event = "quit"
	}
	msgraw, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return msgraw, nil
}

type ClientData2 struct {
	Name      string  `json:"name"`
	Selection Ranges2 `json:"selection"`
}

// DocData2 is structure for document data
type DocData2 struct {
	// Clients  map[string]ClientInfo2 `json:"clients"`
	Clients  map[string]ClientData2 `json:"clients"`
	Document string                 `json:"document"`
	Revision int                    `json:"revision"`
	// Owner      string                 `json:"owner"`
	// Permission int                    `json:"permission"`
	// Editable   bool                   `json:"editable"`
}

// OT session management
// OTClient2 is structure for client connection
type OTClient2 struct {
	conn *websocket.Conn
	sess *OTSession2

	response chan OTResponse
	ClientID string
	UserInfo struct {
		UUID string
		Name string
	}
	Selection []SelData2
}

func NewOTClient2(conn *websocket.Conn, uuid string, name string) *OTClient2 {
	otc := OTClient2{}
	otc.conn = conn
	otc.response = make(chan OTResponse)
	otc.UserInfo.UUID = uuid
	otc.UserInfo.Name = name
	otc.Selection = []SelData2{}
	return &otc
}

// OTClient2 is structure for OT session
type OTSession2 struct {
	db                 *db.DB
	incnum             int
	saveRequest        chan bool
	isSaveTimerRunning bool
	lastUpdater        string

	DocID   string
	DocInfo db.Document
	Clients map[string]*OTClient2
	request chan OTRequest
	OT      *ot.OT
}

var (
	otSessions     = map[string]*OTSession2{}
	otSessionsLock = make(chan bool, 1)
)

func OpenOTSession2(db *db.DB, docID string) (*OTSession2, error) {
	otSessionsLock <- true
	defer func() { <-otSessionsLock }()
	if v, ok := otSessions[docID]; ok {
		return v, nil
	}
	ots := &OTSession2{}
	otSessions[docID] = ots
	ots.db = db
	ots.incnum = 0
	ots.saveRequest = make(chan bool)

	ots.DocID = docID
	docInfo, err := db.GetDocumentInfo(docID)
	if err != nil {
		return nil, err
	}
	ots.DocInfo = docInfo
	ots.Clients = map[string]*OTClient2{}
	ots.request = make(chan OTRequest)

	// TODO: restore OT
	text, err := db.GetLatestDocument(docID)
	if err != nil {
		return nil, err
	}
	ots.OT = ot.New(text)

	go ots.SessionLoop()

	return ots, nil
}

func (sess *OTSession2) Close() {
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

func (sess *OTSession2) SessionLoop() {
	for {
		select {
		case req := <-sess.request:
			if req.Type == OTReqResTypeJoin {
				if v, ok := req.Data.(*OTClient2); ok {
					v.sess = sess
					sess.incnum++
					v.ClientID = strconv.Itoa(sess.incnum)
					sess.Clients[v.ClientID] = v
					v.Response(WSMsgTypeOK, nil)
				}
			} else if req.Type == WSMsgTypeDoc {
				res := DocData2{Clients: map[string]ClientData2{}, Document: sess.OT.Text, Revision: sess.OT.Revision}
				for _, cl := range sess.Clients {
					if cl.ClientID == req.ClientID {
						continue
					}
					rescl := ClientData2{Name: cl.UserInfo.Name}
					rescl.Selection.Ranges = []SelData2{}
					for _, sel := range cl.Selection {
						rescl.Selection.Ranges = append(rescl.Selection.Ranges, sel)
					}
					res.Clients[cl.ClientID] = rescl
				}
				go sess.Clients[req.ClientID].Response(WSMsgTypeDoc, res)
			} else if req.Type == WSMsgTypeOp {
				opdat, ok := req.Data.(OpData2)
				if !ok {
					continue
				}
				ops := ot.Ops{User: req.ClientID, Ops: []ot.Op{}}
				for _, op := range opdat.Operation {
					switch opt := op.(type) {
					case float64:
						opi := int(opt)
						if opi < 0 {
							ops.Ops = append(ops.Ops, ot.Op{OpType: ot.OpTypeDelete, Len: -opi})
						} else {
							ops.Ops = append(ops.Ops, ot.Op{OpType: ot.OpTypeRetain, Len: opi})
						}
					case string:
						ops.Ops = append(ops.Ops, ot.Op{OpType: ot.OpTypeInsert, Len: len(utf16.Encode([]rune(opt))), Text: opt})
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
					if v.OpType == ot.OpTypeRetain {
						opraw = append(opraw, v.Len)
					} else if v.OpType == ot.OpTypeInsert {
						opraw = append(opraw, v.Text)
					} else if v.OpType == ot.OpTypeDelete {
						opraw = append(opraw, -v.Len)
					}
				}
				opdat.Operation = opraw

				sess.Clients[req.ClientID].Selection = opdat.Selection.Ranges

				opres := []interface{}{req.ClientID, opdat.Operation, opdat.Selection}
				for cid, v := range sess.Clients {
					if cid == req.ClientID {
						go v.Response(WSMsgTypeOK, nil)
						continue
					}
					go v.Response(WSMsgTypeOp, opres)
				}

				sess.lastUpdater = sess.Clients[req.ClientID].UserInfo.UUID

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
					time.Sleep(time.Second * 10)
					if !sess.isSaveTimerRunning {
						return
					}
					sess.isSaveTimerRunning = false
					sess.saveRequest <- true
				}()
			} else if req.Type == WSMsgTypeSel {
				seldat, ok := req.Data.(Ranges2)
				if !ok {
					continue
				}
				selresdat := Ranges2{}
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

func (cl *OTClient2) ClientLoop() {
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
				opdat, ok := dat.(OpData2)
				if !ok {
					panic("Logic error")
				}
				cl.sess.Request(WSMsgTypeOp, cl.ClientID, opdat)
			} else if mtype == WSMsgTypeSel {
				opdat, ok := dat.(Ranges2)
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
func (cl *OTSession2) Request(t WSMsgType, cid string, dat interface{}) {
	cl.request <- OTRequest{Type: t, ClientID: cid, Data: dat}
}
func (cl *OTClient2) Response(t WSMsgType, dat interface{}) {
	cl.response <- OTResponse{Type: t, Data: dat}
}
func (sess *OTSession2) AddClient(cl *OTClient2) {
	sess.Request(OTReqResTypeJoin, "", cl)
	<-cl.response
}

func (h *Handler) getOTHandler2(c *gin.Context) {
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Check permission
	docID := c.Param("docid")
	docInfo, err := h.db.GetDocumentInfo(docID)
	if err != nil {
		if err == db.ErrDocumentNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !isRelatedUUID(c, docInfo.OwnerUUID) && docInfo.Permission == db.FilePermPrivate {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	editable := isRelatedUUID(c, docInfo.OwnerUUID) || docInfo.Permission == db.FilePermReadWrite
	if !editable {
		// TODO: support readonly OT
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Setup websocket
	var wsupgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			//TODO
			return true
		},
	}
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v\n", err)
		return
	}
	defer func() { conn.Close() }()

	p, err := h.db.GetProfileByUUID(uuid)
	if err != nil {
		log.Printf("OT handler error: %v", err)
		return
	}
	name := p.Name

	sess, err := OpenOTSession2(h.db, docID)
	if err != nil {
		log.Printf("OT handler error: %v", err)
		return
	}
	otc := NewOTClient2(conn, uuid, name)
	defer func() { close(otc.response) }()
	sess.AddClient(otc)
	sess.Request(WSMsgTypeDoc, otc.ClientID, nil)

	otc.ClientLoop()

	// TODO list
	// OT garbage collection
}

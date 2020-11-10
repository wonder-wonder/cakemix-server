package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	WSMsgTypeOK
	WSMsgTypeOp
	WSMsgTypeSel
	WSMsgTypeQuit
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
	Selection RangesReq2
}

// SelData2 is structure for selection data
type SelData2 struct {
	Anchor int `json:"anchor"`
	Head   int `json:"head"`
}
type RangesReq2 struct {
	Ranges []SelData2 `json:"ranges"`
}
type RangesRes2 struct {
	Ranges map[string][]SelData2 `json:"ranges"`
}

func parseMsg(rawmsg []byte) (WSMsgType, interface{}, error) {
	msg := WSMsg{}
	err := json.Unmarshal(rawmsg, &msg)
	if err != nil {
		return WSMsgTypeUnknown, nil, ErrorInvalidWSMsg
	}
	if msg.Event == "op" {
		dat := OpData2{}
		fmt.Println(string(msg.Data))

		// Separate revision, ops, selections
		var rawdat []json.RawMessage
		err := json.Unmarshal(msg.Data, &rawdat)
		if err != nil {
			println(1)
			return WSMsgTypeUnknown, nil, ErrorInvalidWSMsg
		}
		if len(rawdat) != 3 {
			println(2)
			return WSMsgTypeUnknown, nil, ErrorInvalidWSMsg
		}

		// Revision
		revfloat := 0.0
		err = json.Unmarshal(rawdat[0], &revfloat)
		if err != nil {
			println(3)
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
		dat := RangesReq2{}
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

type ClientData2 struct {
	Name      string     `json:"name"`
	Selection RangesReq2 `json:"selection"`
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

	response chan interface{}
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
	otc.response = make(chan interface{})
	otc.UserInfo.UUID = uuid
	otc.UserInfo.Name = name
	otc.Selection = []SelData2{}
	return &otc
}

// OTClient2 is structure for OT session
type OTSession2 struct {
	db     *db.DB
	incnum int

	DocID   string
	DocInfo db.Document
	Clients map[string]*OTClient2
	request chan OTRequest
	OT      ot.OT
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

	ots.DocID = docID
	docInfo, err := db.GetDocumentInfo(docID)
	if err != nil {
		return nil, err
	}
	ots.DocInfo = docInfo
	ots.Clients = map[string]*OTClient2{}
	ots.request = make(chan OTRequest)

	// TODO: restore OT
	ots.OT.Text = "hoge"

	go ots.SessionLoop()

	return ots, nil
}

type OTReqType int

const (
	OTReqTypeJoin OTReqType = iota
	OTReqTypeDoc
)

type OTRequest struct {
	Type     OTReqType
	ClientID string
	Data     interface{}
}

func (sess *OTSession2) SessionLoop() {
	for {
		select {
		case req := <-sess.request:
			if req.Type == OTReqTypeJoin {
				if v, ok := req.Data.(*OTClient2); ok {
					v.sess = sess
					sess.incnum++
					v.ClientID = strconv.Itoa(sess.incnum)
					sess.Clients[v.ClientID] = v
					v.Response(false)
				}
			} else if req.Type == OTReqTypeDoc {
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
				go sess.Clients[req.ClientID].Response(res)
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
				panic("cclosed")
			}
			mtype, dat, err := parseMsg(req)
			if err != nil {
				//TODO
				fmt.Printf("%v:%v\n", dat)
				panic(err)
			}
			fmt.Printf("%v: %v\n", mtype, dat)
		case resdat := <-cl.response:
			fmt.Printf("res: %v\n", resdat)

			resdatraw, err := json.Marshal(resdat)
			if err != nil {
				//TODO
				panic(err)
			}
			resraw, err := json.Marshal(WSMsg2{Event: "doc", Data: resdatraw})
			if err != nil {
				//TODO
				panic(err)
			}

			cl.conn.WriteMessage(websocket.TextMessage, resraw)
		}
	}
}
func (cl *OTSession2) Request(t OTReqType, cid string, dat interface{}) {
	cl.request <- OTRequest{Type: t, ClientID: cid, Data: dat}
}
func (cl *OTClient2) Response(dat interface{}) {
	cl.response <- dat
}
func (sess *OTSession2) AddClient(cl *OTClient2) {
	sess.Request(OTReqTypeJoin, "", cl)
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

	sess.AddClient(otc)
	sess.Request(OTReqTypeDoc, otc.ClientID, nil)

	otc.ClientLoop()
	println(editable)
	// sess, err := h.getSession(did)
	// if err != nil {
	// 	log.Printf("OT handler error: %v", err)
	// 	return
	// }
	// // Send current session status
	// rev := sess.OT.Revision
	// docdatraw, err := json.Marshal(DocData{Clients: sess.Clinets, Document: sess.OT.Text, Revision: rev, Owner: dinfo.OwnerUUID, Permission: int(dinfo.Permission), Editable: editable})
	// if err != nil {
	// 	log.Printf("OT handler error: %v", err)
	// 	return
	// }
	// initDocRaw, err := json.Marshal(WSMsg{Event: "doc", Data: docdatraw})
	// if err != nil {
	// 	log.Printf("OT handler error: %v", err)
	// 	return
	// }
	// conn.WriteMessage(websocket.TextMessage, initDocRaw)

	// // Add client to session
	// userid := sess.GetNewUserID()
	// sess.AddClient(ClientInfo{Conn: conn, Name: name, ID: userid, UUID: uuid, LastRev: rev})
	// for {
	// 	_, msg, err := conn.ReadMessage()
	// 	if err != nil {
	// 		if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
	// 			break
	// 		}
	// 		log.Printf("OT handler error: %v", err)
	// 		break
	// 	}

	// 	if !editable {
	// 		log.Printf("OT handler error: client tries to write on readonly document.")
	// 		break
	// 	}

	// 	dat := WSMsg{}
	// 	err = json.Unmarshal(msg, &dat)
	// 	if err != nil {
	// 		log.Printf("OT handler error: %v", err)
	// 		break
	// 	}
	// 	if dat.Event == "op" {
	// 		opdat, err := ParseOp(dat.Data, userid)
	// 		if err != nil {
	// 			log.Printf("OT handler error: %v", err)
	// 			break
	// 		}

	// 		op, err := sess.OT.Operate(opdat.Revision, opdat.Operation)
	// 		opraw := []interface{}{}
	// 		for _, v := range op.Ops {
	// 			if v.OpType == ot.OpTypeRetain {
	// 				opraw = append(opraw, v.Len)
	// 			} else if v.OpType == ot.OpTypeInsert {
	// 				opraw = append(opraw, v.Text)
	// 			} else if v.OpType == ot.OpTypeDelete {
	// 				opraw = append(opraw, -v.Len)
	// 			}
	// 		}
	// 		opres := []interface{}{op.User, opraw, map[string][]SelData{"ranges": opdat.Selection}}
	// 		opresraw, err := json.Marshal(opres)
	// 		if err != nil {
	// 			log.Printf("OT handler error: %v", err)
	// 			break
	// 		}
	// 		dat.Data = opresraw
	// 		datraw, err := json.Marshal(dat)
	// 		if err != nil {
	// 			log.Printf("OT handler error: %v", err)
	// 			break
	// 		}
	// 		sess.Broadcast(userid, datraw)

	// 		sess.LastUpdater = uuid
	// 		sess.SaveTimer(h)
	// 		cl, _ := sess.Clinets[userid]
	// 		cl.LastRev = opdat.Revision
	// 		sess.Clinets[userid] = cl

	// 		res := WSMsg{Event: "ok"}
	// 		resraw, err := json.Marshal(res)
	// 		if err != nil {
	// 			log.Printf("OT handler error: %v", err)
	// 			break
	// 		}
	// 		conn.WriteMessage(websocket.TextMessage, resraw)
	// 	} else if dat.Event == "sel" {
	// 		sel, err := ParseSel(dat.Data)
	// 		if err != nil {
	// 			log.Printf("OT handler error: %v", err)
	// 			break
	// 		}
	// 		cl, _ := sess.Clinets[userid]
	// 		cl.Selection = sel
	// 		sess.Clinets[userid] = cl

	// 		selres := []interface{}{userid, map[string][]SelData{"ranges": sel}}
	// 		selresraw, err := json.Marshal(selres)
	// 		if err != nil {
	// 			log.Printf("OT handler error: %v", err)
	// 			break
	// 		}
	// 		dat.Data = selresraw
	// 		datraw, err := json.Marshal(dat)
	// 		if err != nil {
	// 			log.Printf("OT handler error: %v", err)
	// 			break
	// 		}
	// 		sess.Broadcast(userid, datraw)
	// 	}
	// }
	// sess.QuitClient(userid)
}

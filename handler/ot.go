package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/wonder-wonder/cakemix-server/db"
	"github.com/wonder-wonder/cakemix-server/ot"
)

const (
	autoSaveInterval  = 60  //Sec
	otHistGCThreshold = 200 //Ops
)

// {
// 	"e":"op",
// 	"d":
// 	[
// 		1,
// 		[60,"d",17],
// 		{"ranges":[
// 			{"anchor":61,"head":61}
// 		]}
// 	]
// }

// WSMsg is structure for websocket message
type WSMsg struct {
	Event string          `json:"e"`
	Data  json.RawMessage `json:"d,omitempty"`
}

// DocData is structure for document data
type DocData struct {
	Clients  interface{} `json:"clients"`
	Document string      `json:"document"`
	Revision int         `json:"revision"`
}

// ClientInfo is structure for client info
type ClientInfo struct {
	Conn      *websocket.Conn `json:"-"`
	ID        string          `json:"-"`
	UUID      string          `json:"-"`
	LastRev   int             `json:"-"`
	Name      string          `json:"name"`
	Selection []SelData       `json:"selection"`
}

// OpData is structure for operation data
type OpData struct {
	Revision  int
	Operation ot.Ops
	Selection []SelData
}

// SelData is structure for selection data
type SelData struct {
	Anchor int `json:"anchor"`
	Head   int `json:"head"`
}

// ParseOp parses raw operation data to OpData
func ParseOp(msg []byte, user string) (OpData, error) {
	dat := []json.RawMessage{}
	err := json.Unmarshal(msg, &dat)
	if err != nil {
		return OpData{}, err
	}
	if len(dat) < 2 {
		return OpData{}, errors.New("Invalid OT operation message: not enough len")
	}

	ret := OpData{}

	//Revision
	err = json.Unmarshal(dat[0], &ret.Revision)
	if err != nil {
		return OpData{}, err
	}

	//Operations
	opsraw := []interface{}{}
	err = json.Unmarshal(dat[1], &opsraw)
	if err != nil {
		return OpData{}, err
	}
	ret.Operation = ot.Ops{Ops: []ot.Op{}, User: user}
	for _, v := range opsraw {
		switch vt := v.(type) {
		case float64:
			vti := int(vt)
			if vti < 0 {
				ret.Operation.Ops = append(ret.Operation.Ops, ot.Op{OpType: ot.OpTypeDelete, Len: -vti})
			} else {
				ret.Operation.Ops = append(ret.Operation.Ops, ot.Op{OpType: ot.OpTypeRetain, Len: vti})
			}
		case string:
			ret.Operation.Ops = append(ret.Operation.Ops, ot.Op{OpType: ot.OpTypeInsert, Len: len([]rune(vt)), Text: vt})
		default:
			return OpData{}, errors.New("Parse op error")
		}
	}

	//Selection
	ret.Selection = []SelData{}
	if len(dat) == 3 {
		ret.Selection, err = ParseSel(dat[2])
		if err != nil {
			return OpData{}, err
		}
	}

	return ret, nil
}

// ParseSel parses raw selection data to SelData
func ParseSel(msg []byte) ([]SelData, error) {
	type Ranges struct {
		Ranges []SelData `json:"ranges"`
	}
	rang := Ranges{}
	err := json.Unmarshal(msg, &rang)
	if err != nil {
		return []SelData{}, err
	}
	return rang.Ranges, nil
}

// Session is structure for OT session
type Session struct {
	UUID         string
	Clinets      map[string]ClientInfo
	OT           *ot.OT
	TotalClients int
	BCCh         chan BCMsg
	AddCh        chan ClientInfo
	QuitCh       chan string
	isTimerOn    bool
	IDLock       chan bool
	LastUpdater  string
}

// BCMsg is structure for broadcast message
type BCMsg struct {
	from string
	msg  []byte
}

// GetNewUserID returns new user ID
func (sess *Session) GetNewUserID() string {
	sess.IDLock <- true
	sess.TotalClients++
	ret := sess.TotalClients
	<-sess.IDLock
	return strconv.Itoa(ret)
}

// SessionLoop is main loop
func (sess *Session) SessionLoop(h *Handler) {
	for {
		select {
		case bcm := <-sess.BCCh:
			for _, c := range sess.Clinets {
				if c.ID == bcm.from {
					continue
				}
				c.Conn.WriteMessage(websocket.TextMessage, bcm.msg)
			}
		case cl := <-sess.AddCh:
			sess.Clinets[cl.ID] = cl
		case userid := <-sess.QuitCh:
			// Remove client to session
			res := WSMsg{Event: "quit", Data: []byte(userid)}
			resraw, err := json.Marshal(res)
			if err != nil {
				log.Printf("OT handler error: %v", err)
				continue
			}
			for _, c := range sess.Clinets {
				if c.ID == userid {
					continue
				}
				c.Conn.WriteMessage(websocket.TextMessage, resraw)
			}
			delete(sess.Clinets, userid)
			if len(sess.Clinets) == 0 {
				close(sess.AddCh)
				close(sess.BCCh)
				close(sess.QuitCh)
				if len(sess.OT.History) > 0 {
					updateruuid := sess.LastUpdater
					err := h.db.SaveDocument(sess.UUID, updateruuid, sess.OT.Text)
					if err != nil {
						log.Printf("OT handler error: %v", err)
						removeSession(sess.UUID)
						return
					}
					err = h.db.UpdateDocument(sess.UUID, updateruuid)
					if err != nil {
						log.Printf("OT handler error: %v", err)
						removeSession(sess.UUID)
						return
					}
				}
				fmt.Printf("Session(%s) closed: Total %d ops, %s\n", sess.UUID, sess.OT.Revision, sess.OT.Text)
				removeSession(sess.UUID)
				return
			}
		}
	}
}

// Broadcast sends message for all clients
func (sess *Session) Broadcast(from string, msg []byte) {
	sess.BCCh <- BCMsg{from: from, msg: msg}
}

// AddClient adds client into OT session
func (sess *Session) AddClient(cl ClientInfo) {
	sess.AddCh <- cl
}

// QuitClient removes client from OT session
func (sess *Session) QuitClient(userid string) {
	sess.QuitCh <- userid
}

// SaveTimer sets auto save timer
func (sess *Session) SaveTimer(h *Handler) {
	if sess.isTimerOn {
		return
	}
	sess.isTimerOn = true
	go func() {
		<-time.After(autoSaveInterval * time.Second)
		sess.isTimerOn = false
		if len(sess.Clinets) == 0 {
			return
		}
		err := h.db.SaveDocument(sess.UUID, sess.LastUpdater, sess.OT.Text)
		if err != nil {
			log.Printf("OT handler error: %v", err)
			return
		}
		fmt.Printf("Auto saved: session(%s), total %d ops, %s\n", sess.UUID, sess.OT.Revision, sess.OT.Text)
		sess.GCOT()
	}()
}

// GCOT is garbage collector of OT history
func (sess *Session) GCOT() {
	if len(sess.OT.History) < otHistGCThreshold {
		return
	}
	min := sess.OT.Revision
	for _, c := range sess.Clinets {
		if c.LastRev < min {
			min = c.LastRev
		}
	}
	for i := sess.OT.Revision - len(sess.OT.History); i < min; i++ {
		delete(sess.OT.History, i)
	}
	fmt.Printf("GC OT history: rev is %d, len is %d\n", sess.OT.Revision, len(sess.OT.History))
}

var sessions = map[string]*Session{}
var lockch = make(chan bool, 1)

func (h *Handler) getSession(docid string) (*Session, error) {
	// Get session
	lockch <- true
	defer func() { <-lockch }()
	sess, ok := sessions[docid]
	if !ok {
		// If unavailable, init session.
		text, err := h.db.GetLatestDocument(docid)
		if err != nil {
			return nil, err
		}
		sess = &Session{UUID: docid, Clinets: map[string]ClientInfo{}, OT: ot.New(text), TotalClients: 0, BCCh: make(chan BCMsg, 1), AddCh: make(chan ClientInfo, 1), QuitCh: make(chan string, 1), IDLock: make(chan bool, 1)}
		sessions[sess.UUID] = sess
		go sess.SessionLoop(h)
	}
	return sess, nil
}
func removeSession(docid string) {
	lockch <- true
	delete(sessions, docid)
	<-lockch
}

func (h *Handler) getOTHandler(c *gin.Context) {
	uuid, ok := getUUID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	did := c.Param("docid")
	dinfo, err := h.db.GetDocumentInfo(did)
	if err != nil {
		if err == db.ErrDocumentNotFound {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !isRelatedUUID(c, dinfo.OwnerUUID) && dinfo.Permission != db.FilePermReadWrite {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Setup websocket
	var wsupgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v\n", err)
		return
	}

	p, err := h.db.GetProfileByUUID(uuid)
	if err != nil {
		log.Printf("OT handler error: %v", err)
		return
	}
	name := p.Name

	sess, err := h.getSession(did)
	if err != nil {
		log.Printf("OT handler error: %v", err)
		return
	}
	// Send current session status
	rev := sess.OT.Revision
	docdatraw, err := json.Marshal(DocData{Clients: sess.Clinets, Document: sess.OT.Text, Revision: rev})
	if err != nil {
		log.Printf("OT handler error: %v", err)
		return
	}
	initDocRaw, err := json.Marshal(WSMsg{Event: "doc", Data: docdatraw})
	if err != nil {
		log.Printf("OT handler error: %v", err)
		return
	}
	conn.WriteMessage(websocket.TextMessage, initDocRaw)

	// Add client to session
	userid := sess.GetNewUserID()
	sess.AddClient(ClientInfo{Conn: conn, Name: name, ID: userid, UUID: uuid, LastRev: rev})
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
				break
			}
			log.Printf("OT handler error: %v", err)
			break
		}

		dat := WSMsg{}
		err = json.Unmarshal(msg, &dat)
		if err != nil {
			log.Printf("OT handler error: %v", err)
			break
		}
		if dat.Event == "op" {
			opdat, err := ParseOp(dat.Data, userid)
			if err != nil {
				log.Printf("OT handler error: %v", err)
				break
			}

			op, err := sess.OT.Operate(opdat.Revision, opdat.Operation)
			opraw := []interface{}{}
			for _, v := range op.Ops {
				if v.OpType == ot.OpTypeRetain {
					opraw = append(opraw, v.Len)
				} else if v.OpType == ot.OpTypeInsert {
					opraw = append(opraw, v.Text)
				} else if v.OpType == ot.OpTypeDelete {
					opraw = append(opraw, -v.Len)
				}
			}
			opres := []interface{}{op.User, opraw, map[string][]SelData{"ranges": opdat.Selection}}
			opresraw, err := json.Marshal(opres)
			if err != nil {
				log.Printf("OT handler error: %v", err)
				break
			}
			dat.Data = opresraw
			datraw, err := json.Marshal(dat)
			if err != nil {
				log.Printf("OT handler error: %v", err)
				break
			}
			sess.Broadcast(userid, datraw)

			sess.LastUpdater = uuid
			sess.SaveTimer(h)
			cl, _ := sess.Clinets[userid]
			cl.LastRev = opdat.Revision
			sess.Clinets[userid] = cl

			res := WSMsg{Event: "ok"}
			resraw, err := json.Marshal(res)
			if err != nil {
				log.Printf("OT handler error: %v", err)
				break
			}
			conn.WriteMessage(websocket.TextMessage, resraw)
		} else if dat.Event == "sel" {
			sel, err := ParseSel(dat.Data)
			if err != nil {
				log.Printf("OT handler error: %v", err)
				break
			}
			cl, _ := sess.Clinets[userid]
			cl.Selection = sel
			sess.Clinets[userid] = cl

			selres := []interface{}{userid, map[string][]SelData{"ranges": sel}}
			selresraw, err := json.Marshal(selres)
			if err != nil {
				log.Printf("OT handler error: %v", err)
				break
			}
			dat.Data = selresraw
			datraw, err := json.Marshal(dat)
			if err != nil {
				log.Printf("OT handler error: %v", err)
				break
			}
			sess.Broadcast(userid, datraw)
		}
	}
	sess.QuitClient(userid)
}

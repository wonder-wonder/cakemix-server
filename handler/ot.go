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
	"github.com/wonder-wonder/cakemix-server/ot"
)

var clients []ClientInfo

type ClientInfo struct {
	Conn      *websocket.Conn `json:"-"`
	ID        int             `json:"-"`
	Name      string          `json:"name"`
	Selection []SelData       `json:"selection"`
}
type WSMsg struct {
	Event string          `json:"e"`
	Data  json.RawMessage `json:"d"`
}
type DocData struct {
	Clients  interface{} `json:"clients"`
	Document string      `json:"document"`
	Revision int         `json:"revision"`
}

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
type OpData struct {
	Revision  int
	Operation ot.Ops
	Selection []SelData
}

type SelData struct {
	Anchor int `json:"anchor"`
	Head   int `json:"head"`
}

func ParseOp(msg []byte, user string) (OpData, error) {
	dat := []json.RawMessage{}
	err := json.Unmarshal(msg, &dat)
	if err != nil {
		panic(err)
	}
	if len(dat) < 2 {
		panic("fasdfa")
	}

	ret := OpData{}

	//Revision
	err = json.Unmarshal(dat[0], &ret.Revision)
	if err != nil {
		panic(err)
	}

	//Operations
	opsraw := []interface{}{}
	err = json.Unmarshal(dat[1], &opsraw)
	if err != nil {
		panic(err)
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
			(errors.New("Parse op error"))
		}
	}

	//Selection
	ret.Selection = []SelData{}
	if len(dat) == 3 {
		ret.Selection, err = ParseSel(dat[2])
		if err != nil {
			panic(err)
		}
	}

	return ret, nil
}

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

func Broadcast(from int, msg []byte) {
	for _, c := range clients {
		if c.ID == from {
			continue
		}
		c.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}

func (h *Handler) OTHandler(r *gin.RouterGroup) {
	otsess := ot.New("Hello, world!")

	r.GET("/ws", func(c *gin.Context) {
		var wsupgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Failed to set websocket upgrade: %v", err)
			return
		}
		cl := map[string]ClientInfo{}
		for i, c := range clients {
			cl[strconv.Itoa(i)] = c
		}
		docdatraw, err := json.Marshal(DocData{Clients: cl, Document: otsess.Text, Revision: len(otsess.History)})
		if err != nil {
			panic(err)
		}
		initDocRaw, err := json.Marshal(WSMsg{Event: "doc", Data: docdatraw})
		if err != nil {
			panic(err)
		}
		conn.WriteMessage(websocket.TextMessage, initDocRaw)
		useridint := len(clients)
		userid := strconv.Itoa(useridint)
		name := "guest"
		if useridint == 0 {
			name = "alice"
		} else if useridint == 1 {
			name = "bob"
		}
		clients = append(clients, ClientInfo{Conn: conn, ID: useridint, Name: name})
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			dat := WSMsg{}
			err = json.Unmarshal(msg, &dat)
			if err != nil {
				panic(err)
			}
			if dat.Event == "op" {
				opdat, err := ParseOp(dat.Data, userid)
				if err != nil {
					panic(err)
				}

				op, err := otsess.Operate(opdat.Revision, opdat.Operation)
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
					panic(err)
				}
				dat.Data = opresraw
				datraw, err := json.Marshal(dat)
				if err != nil {
					panic(err)
				}
				Broadcast(useridint, datraw)
				fmt.Printf("\n%s\n", otsess.Text)
				temp := `{"e":"ok"}`
				conn.WriteMessage(websocket.TextMessage, []byte(temp))
			} else if dat.Event == "sel" {
				sel, err := ParseSel(dat.Data)
				if err != nil {
					panic(err)
				}
				clients[useridint].Selection = sel

				selres := []interface{}{userid, map[string][]SelData{"ranges": sel}}
				selresraw, err := json.Marshal(selres)
				if err != nil {
					panic(err)
				}
				dat.Data = selresraw
				datraw, err := json.Marshal(dat)
				if err != nil {
					panic(err)
				}
				Broadcast(useridint, datraw)
			}

		}

	})

}

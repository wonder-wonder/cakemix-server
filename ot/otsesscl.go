package ot

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

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
	go func() {
		for !cancel {
			cl.Response(OTReqResTypePing, nil)
			time.Sleep(time.Second * otClientPingInterval)
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
				cl.sess.Request(WSMsgTypeQuit, cl.ClientID, nil)
				return
			}
			if mtype == WSMsgTypeOp {
				opdat, ok := dat.(OpData)
				if !ok {
					log.Printf("OT client error: invalid request data\n")
					cl.sess.Request(WSMsgTypeQuit, cl.ClientID, nil)
					return
				}
				cl.sess.Request(WSMsgTypeOp, cl.ClientID, opdat)
			} else if mtype == WSMsgTypeSel {
				opdat, ok := dat.(Ranges)
				if !ok {
					log.Printf("OT client error: invalid request data\n")
					cl.sess.Request(WSMsgTypeQuit, cl.ClientID, nil)
					return
				}
				cl.sess.Request(WSMsgTypeSel, cl.ClientID, opdat)
			}
		case resdat, ok := <-cl.response:
			if !ok {
				return
			}
			if resdat.Type == OTReqResTypePing {
				cl.conn.WriteMessage(websocket.PingMessage, []byte{})
			} else {
				resraw, err := convertToMsg(resdat.Type, resdat.Data)
				if err != nil {
					log.Printf("OT client error: response error: %v\n", err)
					cl.sess.Request(WSMsgTypeQuit, cl.ClientID, nil)
					return
				}
				cl.conn.WriteMessage(websocket.TextMessage, resraw)
			}
		}
	}
}

// Response responds to client
func (cl *Client) Response(t WSMsgType, dat interface{}) {
	defer func() {
		recover()
	}()
	cl.response <- Response{Type: t, Data: dat}
}

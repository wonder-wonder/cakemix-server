package ot

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

// NewOTClient generates OT client data
func NewOTClient(conn *websocket.Conn, uuid string, name string, iconuri string, readOnly bool) *Client {
	otc := Client{}
	otc.conn = conn
	otc.readOnly = readOnly

	otc.response = make(chan Response)
	otc.UserInfo.UUID = uuid
	otc.UserInfo.Name = name
	otc.UserInfo.IconURI = iconuri
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
				err := cl.conn.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					log.Printf("OT client error: websocket error: %v\n", err)
					cl.sess.Request(WSMsgTypeQuit, cl.ClientID, nil)
					return
				}
			} else {
				resraw, err := convertToMsg(resdat.Type, resdat.Data)
				if err != nil {
					log.Printf("OT client error: response error: %v\n", err)
					cl.sess.Request(WSMsgTypeQuit, cl.ClientID, nil)
					return
				}
				err = cl.conn.WriteMessage(websocket.TextMessage, resraw)
				if err != nil {
					log.Printf("OT client error: websocket error: %v\n", err)
					cl.sess.Request(WSMsgTypeQuit, cl.ClientID, nil)
					return
				}
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

type SessionClient struct {
	// Connection
	conn *websocket.Conn
	// OT status
	clientID  string
	lastRev   int
	selection []SelData
	// User info
	profile  SessionClientProfile
	readOnly bool
	// Server
	cl2sv chan SessionC2SMessage
	sv2cl chan SessionS2CMessage
}

type SessionClientProfile struct {
	UUID    string
	Name    string
	IconURI string
}

func NewSessionClient(conn *websocket.Conn, profile SessionClientProfile, readOnly bool) (*SessionClient, error) {
	cl := &SessionClient{
		conn:      conn,
		clientID:  "",
		lastRev:   0,
		selection: []SelData{},
		profile:   profile,
		readOnly:  readOnly,
		cl2sv:     nil,
		sv2cl:     make(chan SessionS2CMessage),
	}
	return cl, nil
}

func (cl *SessionClient) SendS2C(msgType SessionS2CMessageType, message interface{}) {
	go func() {
		cl.sv2cl <- SessionS2CMessage{
			msgType: msgType,
			message: message,
		}
	}()
}
func (cl *SessionClient) SendC2S(msgType SessionC2SMessageType, message interface{}) {
	go func() {
		cl.cl2sv <- SessionC2SMessage{
			clientID: cl.clientID,
			msgType:  msgType,
			message:  message,
		}
	}()
}

func (cl *SessionClient) Loop() {
	request := make(chan []byte)
	// Reader routine
	ctx := context.Background()
	childCtx, cancel := context.WithCancel(ctx)
	// cancel := false
	// defer func() { cancel = true }()
	defer func() { cancel() }()

	go func() {
		for {
			select {
			case <-childCtx.Done():
				err := cl.conn.Close()
				if err != nil {
					log.Printf("OT client error: ws close error: %v\n", err)
					close(request)
				}
				return
			default:
				_, msg, err := cl.conn.ReadMessage()
				if err != nil {
					if operr, ok := err.(*net.OpError); ok && operr.Timeout() {
						continue
					}
					// Closed
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
						close(request)
						return
					}
					log.Printf("OT client error: read error: %v\n", err)
					close(request)
					err := cl.conn.Close()
					if err != nil {
						log.Printf("OT client error: ws close error: %v\n", err)
						close(request)
					}
					return
				}
				request <- msg
			}
		}

	}()
	go func() {
		for {
			select {
			case <-childCtx.Done():
				return
			default:
				cl.SendS2C(SessionS2CMessageTypePing, nil)
				time.Sleep(time.Second * otClientPingInterval)
			}
		}
	}()
	for {
		select {
		case s2cmsg, ok := <-cl.sv2cl:
			if !ok {
				// Closed by server
				return
			}
			switch s2cmsg.msgType {
			case SessionS2CMessageTypePing:
				err := cl.conn.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					log.Printf("OT client error: websocket error: %v\n", err)
					cl.Stop()
					return
				}
			case SessionS2CMessageTypeWSMsg:
				wsmsg := s2cmsg.message.(SessionWSMessage)
				resraw, err := convertToMsg(wsmsg.Event, wsmsg.Data)
				if err != nil {
					log.Printf("OT client error: response error: %v\n", err)
					cl.Stop()
					return
				}
				err = cl.conn.WriteMessage(websocket.TextMessage, resraw)
				if err != nil {
					log.Printf("OT client error: websocket error: %v\n", err)
					cl.Stop()
					return
				}
			}
		case req, ok := <-request:
			if !ok {
				cl.Stop()
				return
			}
			if cl.readOnly {
				log.Printf("OT client error: permission denied\n")
				cl.Stop()
				return
			}
			mtype, dat, err := parseMsg(req)
			if err != nil {
				log.Printf("OT client error: %v\n", err)
				cl.Stop()
				return
			}
			if mtype == WSMsgTypeOp {
				opdat, ok := dat.(OpData)
				if !ok {
					log.Printf("OT client error: invalid request data\n")
					cl.Stop()
					return
				}
				cl.SendC2S(SessionC2SMessageTypeWSMsg, SessionWSMessage{Event: WSMsgTypeOp, Data: opdat})
			} else if mtype == WSMsgTypeSel {
				opdat, ok := dat.(Ranges)
				if !ok {
					log.Printf("OT client error: invalid request data\n")
					cl.Stop()
					return
				}
				cl.SendC2S(SessionC2SMessageTypeWSMsg, SessionWSMessage{Event: WSMsgTypeSel, Data: opdat})
			}
		}
	}

}
func (cl *SessionClient) Stop() {
	// Closed by client
	cl.SendC2S(SessionC2SMessageTypeClose, nil)
	// Wait server close
	_, ok := <-cl.sv2cl
	for ok {
		_, ok = <-cl.sv2cl
	}
}

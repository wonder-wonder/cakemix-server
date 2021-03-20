package ot

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

// Client is structure for client connection
type Client struct {
	// Connection
	conn *websocket.Conn
	// OT status
	clientID  string
	lastRev   int
	selection []SelData
	// User info
	profile  ClientProfile
	readOnly bool
	// Server
	cl2sv chan otC2SMessage
	sv2cl chan otS2CMessage
}

// ClientProfile is structure for client profile
type ClientProfile struct {
	UUID    string
	Name    string
	IconURI string
}

// NewClient generates OTClient
func NewClient(conn *websocket.Conn, profile ClientProfile, readOnly bool) (*Client, error) {
	cl := &Client{
		conn:      conn,
		clientID:  "",
		lastRev:   0,
		selection: []SelData{},
		profile:   profile,
		readOnly:  readOnly,
		cl2sv:     nil,
		sv2cl:     make(chan otS2CMessage, 100),
	}
	return cl, nil
}

func (cl *Client) sendS2C(msgType otS2CMessageType, message interface{}) {
	go func() {
		cl.sv2cl <- otS2CMessage{
			msgType: msgType,
			message: message,
		}
	}()
}
func (cl *Client) sendC2S(msgType otC2SMessageType, message interface{}) {
	go func() {
		cl.cl2sv <- otC2SMessage{
			clientID: cl.clientID,
			msgType:  msgType,
			message:  message,
		}
	}()
}

// Loop is main loop for client
func (cl *Client) Loop() {
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
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure) {
						close(request)
						return
					}
					log.Printf("OT client error: read error: %v\n", err)
					close(request)
					err := cl.conn.Close()
					if err != nil {
						log.Printf("OT client error: ws close error: %v\n", err)
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
				cl.sendS2C(otS2CMessageTypePing, nil)
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
			case otS2CMessageTypePing:
				err := cl.conn.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					log.Printf("OT client error: websocket error: %v\n", err)
					cl.stop()
					return
				}
			case otS2CMessageTypeWSMsg:
				wsmsg := s2cmsg.message.(otWSMessage)
				resraw, err := convertToMsg(wsmsg.Event, wsmsg.Data)
				if err != nil {
					log.Printf("OT client error: response error: %v\n", err)
					cl.stop()
					return
				}
				err = cl.conn.WriteMessage(websocket.TextMessage, resraw)
				if err != nil {
					log.Printf("OT client error: websocket error: %v\n", err)
					cl.stop()
					return
				}
			}
		case req, ok := <-request:
			if !ok {
				cl.stop()
				return
			}
			if cl.readOnly {
				log.Printf("OT client error: permission denied\n")
				cl.stop()
				return
			}
			mtype, dat, err := parseMsg(req)
			if err != nil {
				log.Printf("OT client error: %v\n", err)
				cl.stop()
				return
			}
			if mtype == WSMsgTypeOp {
				opdat, ok := dat.(OpData)
				if !ok {
					log.Printf("OT client error: invalid request data\n")
					cl.stop()
					return
				}
				cl.sendC2S(otC2SMessageTypeWSMsg, otWSMessage{Event: WSMsgTypeOp, Data: opdat})
			} else if mtype == WSMsgTypeSel {
				opdat, ok := dat.(Ranges)
				if !ok {
					log.Printf("OT client error: invalid request data\n")
					cl.stop()
					return
				}
				cl.sendC2S(otC2SMessageTypeWSMsg, otWSMessage{Event: WSMsgTypeSel, Data: opdat})
			}
		}
	}

}
func (cl *Client) stop() {
	// Closed by client
	cl.sendC2S(otC2SMessageTypeClose, nil)
	// Wait server close
	_, ok := <-cl.sv2cl
	for ok {
		_, ok = <-cl.sv2cl
	}
}

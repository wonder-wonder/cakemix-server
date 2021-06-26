package ot

import (
	"log"
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
	sv2cl chan otWSMessage
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
		sv2cl:     make(chan otWSMessage, 1000),
	}
	return cl, nil
}

func (cl *Client) sendC2S(msgType otC2SMessageType, message interface{}) {
	cl.cl2sv <- otC2SMessage{
		clientID: cl.clientID,
		msgType:  msgType,
		message:  message,
	}
}

// Loop is main loop for client
func (cl *Client) Loop() {
	request := make(chan []byte)

	// Ping timer
	pingTicker := time.NewTicker(time.Second * otClientPingInterval)
	defer pingTicker.Stop()

	// Reader routine
	readstop := make(chan struct{})
	defer func() { close(readstop) }()
	go func() {
		defer func() { close(request) }()
		for {
			select {
			case <-readstop:
				err := cl.conn.Close()
				if err != nil {
					log.Printf("OT client error: ws close error: %v\n", err)
				}
				return
			default:
				_, msg, err := cl.conn.ReadMessage()
				if err != nil {
					// Closed
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure) {
						return
					}
					log.Printf("OT client error: read error: %v\n", err)
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

	sendSvResponse := func(s2cmsg otWSMessage) bool {
		resraw, err := convertToMsg(s2cmsg.Event, s2cmsg.Data)
		if err != nil {
			log.Printf("OT client error: response error: %v\n", err)
			return false
		}
		err = cl.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
		if err != nil {
			log.Printf("OT client error: websockest error: %v\n", err)
			return false
		}
		err = cl.conn.WriteMessage(websocket.TextMessage, resraw)
		if err != nil {
			log.Printf("OT client error: websocket error: %v\n", err)
			return false
		}
		err = cl.conn.SetWriteDeadline(time.Time{})
		if err != nil {
			log.Printf("OT client error: websockest error: %v\n", err)
			return false
		}
		return true
	}

main:
	for {
		select {
		case s2cmsg, ok := <-cl.sv2cl:
			// Server response is high priority so check the first
			if !ok {
				// Closed by server and notification is not needed.
				return
			}
			if !sendSvResponse(s2cmsg) {
				break main
			}
		default:
		}
		// If no server response, check all response including server response
		select {
		case <-pingTicker.C:
			err := cl.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err != nil {
				log.Printf("OT client error: websockest error: %v\n", err)
				break main
			}
			err = cl.conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Printf("OT client error: websocket error: %v\n", err)
				break main
			}
			err = cl.conn.SetWriteDeadline(time.Time{})
			if err != nil {
				log.Printf("OT client error: websockest error: %v\n", err)
				break main
			}
		case s2cmsg, ok := <-cl.sv2cl:
			if !ok {
				// Closed by server and notification is not needed.
				return
			}
			if !sendSvResponse(s2cmsg) {
				break main
			}
		case req, ok := <-request:
			if !ok {
				break main
			}
			if cl.readOnly {
				log.Printf("OT client error: permission denied\n")
				break main
			}
			mtype, dat, err := parseMsg(req)
			if err != nil {
				log.Printf("OT client error: %v\n", err)
				break main
			}
			if mtype == WSMsgTypeOp {
				opdat, ok := dat.(OpData)
				if !ok {
					log.Printf("OT client error: invalid request data\n")
					break main
				}
				cl.sendC2S(otC2SMessageTypeWSMsg, otWSMessage{Event: WSMsgTypeOp, Data: opdat})
			} else if mtype == WSMsgTypeSel {
				opdat, ok := dat.(Ranges)
				if !ok {
					log.Printf("OT client error: invalid request data\n")
					break main
				}
				cl.sendC2S(otC2SMessageTypeWSMsg, otWSMessage{Event: WSMsgTypeSel, Data: opdat})
			}
		}
	}
	// Discard all server response data
	clear := false
	for !clear {
		select {
		case _, ok := <-cl.sv2cl:
			if !ok {
				clear = true
			}
		default:
			clear = true
		}
	}
	// Closed by client
	cl.sendC2S(otC2SMessageTypeClose, nil)
	// Wait server close
	_, ok := <-cl.sv2cl
	for ok {
		_, ok = <-cl.sv2cl
	}
}

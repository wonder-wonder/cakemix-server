package ot

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type OTClient struct {
	// Connection
	conn *websocket.Conn
	// OT status
	clientID  string
	lastRev   int
	selection []SelData
	// User info
	profile  OTClientProfile
	readOnly bool
	// Server
	cl2sv chan OTC2SMessage
	sv2cl chan OTS2CMessage
}

type OTClientProfile struct {
	UUID    string
	Name    string
	IconURI string
}

func NewOTClient(conn *websocket.Conn, profile OTClientProfile, readOnly bool) (*OTClient, error) {
	cl := &OTClient{
		conn:      conn,
		clientID:  "",
		lastRev:   0,
		selection: []SelData{},
		profile:   profile,
		readOnly:  readOnly,
		cl2sv:     nil,
		sv2cl:     make(chan OTS2CMessage),
	}
	return cl, nil
}

func (cl *OTClient) SendS2C(msgType OTS2CMessageType, message interface{}) {
	go func() {
		cl.sv2cl <- OTS2CMessage{
			msgType: msgType,
			message: message,
		}
	}()
}
func (cl *OTClient) SendC2S(msgType OTC2SMessageType, message interface{}) {
	go func() {
		cl.cl2sv <- OTC2SMessage{
			clientID: cl.clientID,
			msgType:  msgType,
			message:  message,
		}
	}()
}

func (cl *OTClient) Loop() {
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
				cl.SendS2C(OTS2CMessageTypePing, nil)
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
			case OTS2CMessageTypePing:
				err := cl.conn.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					log.Printf("OT client error: websocket error: %v\n", err)
					cl.Stop()
					return
				}
			case OTS2CMessageTypeWSMsg:
				wsmsg := s2cmsg.message.(OTWSMessage)
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
				cl.SendC2S(OTC2SMessageTypeWSMsg, OTWSMessage{Event: WSMsgTypeOp, Data: opdat})
			} else if mtype == WSMsgTypeSel {
				opdat, ok := dat.(Ranges)
				if !ok {
					log.Printf("OT client error: invalid request data\n")
					cl.Stop()
					return
				}
				cl.SendC2S(OTC2SMessageTypeWSMsg, OTWSMessage{Event: WSMsgTypeSel, Data: opdat})
			}
		}
	}

}
func (cl *OTClient) Stop() {
	// Closed by client
	cl.SendC2S(OTC2SMessageTypeClose, nil)
	// Wait server close
	_, ok := <-cl.sv2cl
	for ok {
		_, ok = <-cl.sv2cl
	}
}

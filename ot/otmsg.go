package ot

import (
	"encoding/json"
	"errors"
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

// WSMsg is structure for websocket message
type WSMsg struct {
	Event string          `json:"e"`
	Data  json.RawMessage `json:"d,omitempty"`
}

// OpData is structure for OT operation data
type OpData struct {
	Revision  int
	Operation []interface{}
	Selection Ranges
}

// SelData is structure for selection data
type SelData struct {
	Anchor int `json:"anchor"`
	Head   int `json:"head"`
}

// Ranges is structure for selection array data
type Ranges struct {
	Ranges []SelData `json:"ranges"`
}

func parseMsg(rawmsg []byte) (WSMsgType, interface{}, error) {
	msg := WSMsg{}
	err := json.Unmarshal(rawmsg, &msg)
	if err != nil {
		return WSMsgTypeUnknown, nil, ErrorInvalidWSMsg
	}
	if msg.Event == "op" {
		dat := OpData{}

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
		dat := Ranges{}
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

// ClientData is structure for client information
type ClientData struct {
	Name      string `json:"name"`
	Selection Ranges `json:"selection"`
}

// DocData is structure for document data
type DocData struct {
	Clients    map[string]ClientData `json:"clients"`
	Document   string                `json:"document"`
	Revision   int                   `json:"revision"`
	Owner      string                `json:"owner"`
	Permission int                   `json:"permission"`
	Editable   bool                  `json:"editable"`
}

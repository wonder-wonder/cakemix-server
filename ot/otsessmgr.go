package ot

import (
	"errors"
	"time"

	"github.com/wonder-wonder/cakemix-server/db"
)

const serverStopDelay = 30

type otStatus int

const (
	otStatusStarting otStatus = iota
	otStatusRunning
	otStatusStopping
)

type otServerRequestType int

const (
	otServerRequestTypeStarted otServerRequestType = iota
	otServerRequestTypeClientClosed
	otServerRequestTypeStopped
)

type otManagerRequestType int

const (
	otManagerRequestTypeAddClient otManagerRequestType = iota
)

// Manager is structure for ot management
type Manager struct {
	db        *db.DB
	sesslist  map[string]*otInfo
	clientReq chan otClientRequest
	serverReq chan otServerRequest
	timeout   chan string
	stop      chan string
}

type otInfo struct {
	Server    *Server
	ClientNum int
	Status    otStatus
	StopTimer *time.Timer
	StopWhen  time.Time
}
type otClientRequest struct {
	ready  chan struct{}
	docID  string
	client *Client
}
type otServerRequest struct {
	docID   string
	reqType otServerRequestType
	request interface{}
}
type otManagerRequest struct {
	reqType otManagerRequestType
	request interface{}
}

// NewManager creates new manager
func NewManager(db *db.DB) (*Manager, error) {
	mgr := &Manager{
		db:        db,
		sesslist:  map[string]*otInfo{},
		clientReq: make(chan otClientRequest),
		serverReq: make(chan otServerRequest),
		timeout:   make(chan string),
		stop:      make(chan string),
	}
	return mgr, nil
}

// Loop is main loop for manager
func (mgr *Manager) Loop() {
	for {
		select {
		case clreq, _ := <-mgr.clientReq:
			svinfo, ok := mgr.sesslist[clreq.docID]
			if !ok {
				go func() {
					err := mgr.StartServer(clreq.docID)
					if err != nil {
						panic(err)
					}
				}()
			}
			if !ok || svinfo.Status != otStatusRunning {
				// Reenqueue
				go func() {
					time.Sleep(time.Millisecond * 10)
					mgr.clientReq <- clreq
				}()
				continue
			}

			mgr2sv := svinfo.Server.mgr2sv
			// If manager request buffer is full,
			if len(mgr2sv)+1 >= cap(mgr2sv) {
				// Reenqueue
				go func() {
					time.Sleep(time.Millisecond * 10)
					mgr.clientReq <- clreq
				}()
				continue
			}
			mgr2sv <- otManagerRequest{
				reqType: otManagerRequestTypeAddClient,
				request: &clreq,
			}

			svinfo.ClientNum++
		case svreq, _ := <-mgr.serverReq:
			switch svreq.reqType {
			case otServerRequestTypeStarted:
				svinfo, ok := mgr.sesslist[svreq.docID]
				if !ok {
					continue
				}
				svinfo.Status = otStatusRunning
			case otServerRequestTypeClientClosed:
				svinfo, ok := mgr.sesslist[svreq.docID]
				if !ok {
					continue
				}
				svinfo.ClientNum--
				if svinfo.ClientNum == 0 {
					if svinfo.Status == otStatusStopping {
						continue
					}
					// Set timeout
					svinfo.StopWhen = time.Now().Add(time.Second * serverStopDelay)
					timer := svinfo.StopTimer
					if timer != nil {
						timer.Stop()
					}
					svinfo.StopTimer = time.AfterFunc(time.Second*30, func() {
						mgr.timeout <- svinfo.Server.docID
					})
				}
			case otServerRequestTypeStopped:
				delete(mgr.sesslist, svreq.docID)
			}
		case docID := <-mgr.timeout:
			svinfo, ok := mgr.sesslist[docID]
			if !ok {
				continue
			}
			if time.Now().Before(svinfo.StopWhen) {
				continue
			}
			if svinfo.ClientNum > 0 {
				continue
			}
			close(svinfo.Server.mgr2sv)
			svinfo.Status = otStatusStopping
		case docID := <-mgr.stop:
			if docID != "" {
				svinfo, ok := mgr.sesslist[docID]
				if !ok {
					continue
				}
				close(svinfo.Server.mgr2sv)
				svinfo.Status = otStatusStopping
				continue
			}
			for _, v := range mgr.sesslist {
				close(v.Server.mgr2sv)
				v.Status = otStatusStopping
			}
			for len(mgr.sesslist) > 0 {
				svreq := <-mgr.serverReq
				if svreq.reqType == otServerRequestTypeStopped {
					delete(mgr.sesslist, svreq.docID)
				}
			}
			return
		}
	}
}

// StartServer creates new server and start main loop
func (mgr *Manager) StartServer(docID string) error {
	if _, ok := mgr.sesslist[docID]; ok {
		return errors.New("Server already exist: " + docID)
	}
	sv, err := NewServer(docID, mgr.serverReq, mgr.db)
	if err != nil {
		return err
	}
	mgr.sesslist[docID] = &otInfo{
		Server:    sv,
		ClientNum: 0,
		Status:    otStatusStarting,
		StopTimer: nil,
	}
	go sv.Loop()
	return nil
}

// ClientConnect connects client to server
func (mgr *Manager) ClientConnect(cl *Client, docid string) {
	ready := make(chan struct{})
	mgr.clientReq <- otClientRequest{
		ready:  ready,
		docID:  docid,
		client: cl,
	}
	<-ready
}

// StopOTManager stops manager
func (mgr *Manager) StopOTManager() {
	mgr.stop <- ""
}

// StopOTSession stops session
func (mgr *Manager) StopOTSession(docID string) {
	go func() { mgr.stop <- docID }()
}

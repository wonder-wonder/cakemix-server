package ot

import (
	"errors"
	"time"

	"github.com/wonder-wonder/cakemix-server/db"
)

const serverStopDelay = 30

type SessionStatus int

const (
	SessionStatusStarting SessionStatus = iota
	SessionStatusRunning
	SessionStatusStopping
)

type SessionServerRequestType int

const (
	SessionServerRequestTypeStarted SessionServerRequestType = iota
	SessionServerRequestTypeClientClosed
	SessionServerRequestTypeStopped
)

type SessionManagerRequestType int

const (
	SessionManagerRequestTypeAddClient SessionManagerRequestType = iota
)

type SessionManager struct {
	db        *db.DB
	sesslist  map[string]*SessionInfo
	clientReq chan SessionClientRequest
	serverReq chan SessionServerRequest
	timeout   chan string
}

type SessionInfo struct {
	Server    *SessionServer
	ClientNum int
	Status    SessionStatus
	StopTimer *time.Timer
	StopWhen  time.Time
}
type SessionClientRequest struct {
	ready  chan struct{}
	docID  string
	client *SessionClient
}
type SessionServerRequest struct {
	docID   string
	reqType SessionServerRequestType
	request interface{}
}
type SessionManagerRequest struct {
	reqType SessionManagerRequestType
	request interface{}
}

func NewSessionManager(db *db.DB) (*SessionManager, error) {
	mgr := &SessionManager{
		db:        db,
		sesslist:  map[string]*SessionInfo{},
		clientReq: make(chan SessionClientRequest),
		serverReq: make(chan SessionServerRequest),
		timeout:   make(chan string),
	}
	return mgr, nil
}

func (mgr *SessionManager) Loop() {
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
			if !ok || svinfo.Status != SessionStatusRunning {
				// Reenqueue
				go func() {
					time.Sleep(time.Millisecond * 10)
					mgr.clientReq <- clreq
				}()
				continue
			}
			svinfo.Server.AddClient(&clreq)
			svinfo.ClientNum++
		case svreq, _ := <-mgr.serverReq:
			switch svreq.reqType {
			case SessionServerRequestTypeStarted:
				svinfo := mgr.sesslist[svreq.docID]
				svinfo.Status = SessionStatusRunning
			case SessionServerRequestTypeClientClosed:
				svinfo := mgr.sesslist[svreq.docID]
				svinfo.ClientNum--
				if svinfo.ClientNum == 0 {
					if svinfo.Status == SessionStatusStopping {
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
			case SessionServerRequestTypeStopped:
				delete(mgr.sesslist, svreq.docID)
			}
		case docID := <-mgr.timeout:
			svinfo := mgr.sesslist[docID]
			if time.Now().Before(svinfo.StopWhen) {
				continue
			}
			if svinfo.ClientNum > 0 {
				continue
			}
			close(svinfo.Server.mgr2sv)
			svinfo.Status = SessionStatusStopping
		}
	}
}

func (mgr *SessionManager) StartServer(docID string) error {
	if _, ok := mgr.sesslist[docID]; ok {
		return errors.New("Server already exist: " + docID)
	}
	sv, err := NewSessionServer(docID, mgr.serverReq, mgr.db)
	if err != nil {
		return err
	}
	mgr.sesslist[docID] = &SessionInfo{
		Server:    sv,
		ClientNum: 0,
		Status:    SessionStatusStarting,
		StopTimer: nil,
	}
	go sv.Loop()
	return nil
}

func (mgr *SessionManager) ClientConnect(cl *SessionClient, docid string) {
	ready := make(chan struct{})
	mgr.clientReq <- SessionClientRequest{
		ready:  ready,
		docID:  docid,
		client: cl,
	}
	<-ready
}

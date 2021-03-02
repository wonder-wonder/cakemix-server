package ot

import (
	"errors"
	"time"

	"github.com/wonder-wonder/cakemix-server/db"
)

const serverStopDelay = 30

type OTStatus int

const (
	OTStatusStarting OTStatus = iota
	OTStatusRunning
	OTStatusStopping
)

type OTServerRequestType int

const (
	OTServerRequestTypeStarted OTServerRequestType = iota
	OTServerRequestTypeClientClosed
	OTServerRequestTypeStopped
)

type OTManagerRequestType int

const (
	OTManagerRequestTypeAddClient OTManagerRequestType = iota
)

type OTManager struct {
	db        *db.DB
	sesslist  map[string]*OTInfo
	clientReq chan OTClientRequest
	serverReq chan OTServerRequest
	timeout   chan string
	stop      chan struct{}
}

type OTInfo struct {
	Server    *OTServer
	ClientNum int
	Status    OTStatus
	StopTimer *time.Timer
	StopWhen  time.Time
}
type OTClientRequest struct {
	ready  chan struct{}
	docID  string
	client *OTClient
}
type OTServerRequest struct {
	docID   string
	reqType OTServerRequestType
	request interface{}
}
type OTManagerRequest struct {
	reqType OTManagerRequestType
	request interface{}
}

func NewOTManager(db *db.DB) (*OTManager, error) {
	mgr := &OTManager{
		db:        db,
		sesslist:  map[string]*OTInfo{},
		clientReq: make(chan OTClientRequest),
		serverReq: make(chan OTServerRequest),
		timeout:   make(chan string),
		stop:      make(chan struct{}),
	}
	return mgr, nil
}

func (mgr *OTManager) Loop() {
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
			if !ok || svinfo.Status != OTStatusRunning {
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
			case OTServerRequestTypeStarted:
				svinfo := mgr.sesslist[svreq.docID]
				svinfo.Status = OTStatusRunning
			case OTServerRequestTypeClientClosed:
				svinfo := mgr.sesslist[svreq.docID]
				svinfo.ClientNum--
				if svinfo.ClientNum == 0 {
					if svinfo.Status == OTStatusStopping {
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
			case OTServerRequestTypeStopped:
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
			svinfo.Status = OTStatusStopping
		case <-mgr.stop:
			for _, v := range mgr.sesslist {
				close(v.Server.mgr2sv)
				v.Status = OTStatusStopping
			}
			for len(mgr.sesslist) > 0 {
				svreq := <-mgr.serverReq
				if svreq.reqType == OTServerRequestTypeStopped {
					delete(mgr.sesslist, svreq.docID)
				}
			}
			return
		}
	}
}

func (mgr *OTManager) StartServer(docID string) error {
	if _, ok := mgr.sesslist[docID]; ok {
		return errors.New("Server already exist: " + docID)
	}
	sv, err := NewOTServer(docID, mgr.serverReq, mgr.db)
	if err != nil {
		return err
	}
	mgr.sesslist[docID] = &OTInfo{
		Server:    sv,
		ClientNum: 0,
		Status:    OTStatusStarting,
		StopTimer: nil,
	}
	go sv.Loop()
	return nil
}

func (mgr *OTManager) ClientConnect(cl *OTClient, docid string) {
	ready := make(chan struct{})
	mgr.clientReq <- OTClientRequest{
		ready:  ready,
		docID:  docid,
		client: cl,
	}
	<-ready
}
func (mgr *OTManager) StopOTManager() {
	mgr.stop <- struct{}{}
}

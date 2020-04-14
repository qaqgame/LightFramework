package fsplite

import (
	"encoding/binary"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// FSPGateway : FSP Gateway
type FSPGateway struct {
	mapSession       map[uint32]*FSPSession
	IsRunning        bool
	Param            *FSPParam

	conn             *net.UDPConn
	receiveBuffer    []byte
	port             int
	closeSignal      chan int
	lastClearSeesionTime  int64

	recvRunning      bool
	logger           *logrus.Entry
	rwMutex          sync.RWMutex
}	

// NewSessionID :
func NewSessionID() uint32 {
	sid++
	return sid
}

// NewFSPGateway :
func NewFSPGateway(_port int) *FSPGateway {
	fspgateway := new(FSPGateway)
	fspgateway.port = _port;

	fspgateway.mapSession = make(map[uint32]*FSPSession)
	fspgateway.receiveBuffer = make([]byte, 4096)

	fspgateway.logger = logrus.WithFields(logrus.Fields{"Process":"FSPGateway"})
	fspgateway.IsRunning = false
	fspgateway.conn = nil
	fspgateway.closeSignal = make(chan int, 2)
	fspgateway.recvRunning = false
	fspgateway.lastClearSeesionTime = 0
	fspgateway.rwMutex = sync.RWMutex{}

	fspgateway.Start()
	return fspgateway
}

// CreateSession :
func (fspgateway *FSPGateway) CreateSession() *FSPSession {
	sid := NewSessionID()
	session := NewFSPSession(sid, fspgateway.HandleSessionSend)
	fspgateway.rwMutex.Lock()
	fspgateway.mapSession[sid] = session
	fspgateway.rwMutex.Unlock()
	return session
}

// ReleaseSession :
func (fspgateway *FSPGateway) ReleaseSession(sid uint32) {
	session := fspgateway.mapSession[sid]
	if session != nil {
		fspgateway.rwMutex.Lock()
		delete(fspgateway.mapSession, sid)
		fspgateway.rwMutex.Unlock()
	}
}

// HandleSessionSend :
func (fspgateway *FSPGateway) HandleSessionSend(endpoint *net.UDPAddr, bytes []byte, lenght int) {
	fspgateway.logger.Debug("HandSessionSend in FSPGateway")
	if fspgateway.conn != nil {
		n, err := fspgateway.conn.WriteToUDP(bytes[:lenght],endpoint)
		fspgateway.logger.Debug("HandleSessionSend In FSPGateway, wirte len: ",n)
		if err != nil {
			fspgateway.logger.Error("HandleSessionSend In FSPGateway, write failed")
		}
	} else {
		fspgateway.logger.Warn("HandleSessionSend in FSPGateway, conn is nil")
	}
}


// Clean : 
func (fspgateway *FSPGateway) Clean() {
	fspgateway.rwMutex.Lock()
	fspgateway.mapSession = make(map[uint32]*FSPSession)
	fspgateway.rwMutex.Unlock()
	fspgateway.Close()
}

// Start :
func (fspgateway *FSPGateway) Start() {
	fspgateway.logger.Debug("FSPGateway Start")
	fspgateway.IsRunning = true

	udpAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0"+strconv.Itoa(fspgateway.port))
	if err != nil {
		fspgateway.logger.Fatal("Start Gateway error(resolveUDPAddr): ",err)
	}

	fspgateway.conn, err = net.ListenUDP("udp4",udpAddr)
	if err != nil {
		fspgateway.logger.Fatal("Start gateway err(listenUDP): ",err)
	}

	fspgateway.logger.Info("Listen UDP: ", fspgateway.port, udpAddr)

	for _,v := range fspgateway.mapSession {
		if !v.ReceiveInMainActive {
			v.FSPSessionInit()
		}
	}

	go fspgateway.Recv()
}

// Close :
func (fspgateway *FSPGateway) Close() {
	fspgateway.IsRunning = false

	for _,v := range fspgateway.mapSession {
		if v.ReceiveInMainActive {
			v.StopReceive()
		}
	}

	// if fspgateway.recvRunning {
	// 	fspgateway.closeSignal <- -1
	// }

	fspgateway.conn.Close()
	fspgateway.conn = nil
}


// Recv : receive goroutine
func (fspgateway *FSPGateway) Recv() {
	fspgateway.logger.Debug("Start Recv Goroutine Of FSPGateway")
	fspgateway.recvRunning = true

	for fspgateway.IsRunning {
		fspgateway.DoReceiveInGoroutine()
	}

	fspgateway.recvRunning = false
}

// GetSession : get Session from Gatway
func (fspgateway *FSPGateway) GetSession(sid uint32) *FSPSession {
	return fspgateway.mapSession[sid]
}

// DoReceiveInGoroutine :
func (fspgateway *FSPGateway) DoReceiveInGoroutine() {
	fspgateway.logger.Debug("In function DoReceiveGoroutine of FSPGateway")
	sidbuf := make([]byte, 4)
	n, addr, err := fspgateway.conn.ReadFromUDP(fspgateway.receiveBuffer)
	if err != nil {
		//log.Println("error DoReceiveInThread err: ",err)
		fspgateway.logger.Error("error DoReceiveInGoroutine of FSPGateway: ", err)
	}
	//log.Println(time.Now().Unix(),"Received data: ", n)
	fspgateway.logger.Debug("Received data from UDP in FSPGateway, length is ", n)
	if n > 0 {
		sidbuf = fspgateway.receiveBuffer[:4]

		var session *FSPSession = nil
		sid := binary.BigEndian.Uint32(sidbuf)
		if sid == 0 {
			fspgateway.logger.Info("Sid 为0，丢弃该包")
		} else {
			session = fspgateway.GetSession(sid)
		}

		if session != nil {
			session.Active(addr)
			session.DoReceiveInGateway(fspgateway.receiveBuffer, n)

		} else {
			fspgateway.logger.Warn("sid 不存在")
		}
	}
}

// Tick :
func (fspgateway *FSPGateway) Tick() {
	if fspgateway.IsRunning {
		discrepancy := uint32(time.Now().Sub(reftime) / time.Millisecond)
		currentTime := time.Now().Unix()
		if currentTime - fspgateway.lastClearSeesionTime > SessionActiveTimeout {
			fspgateway.lastClearSeesionTime = currentTime
			fspgateway.ClearNoActiveSession()
		}

		for _,v := range fspgateway.mapSession {
			v.Tick(discrepancy)
		}
	}
}

// ClearNoActiveSession :
func (fspgateway *FSPGateway) ClearNoActiveSession() {
	for k,v := range fspgateway.mapSession {
		// clear
		if !v.IsActive() {
			fspgateway.rwMutex.Lock()
			delete(fspgateway.mapSession, k)
			fspgateway.rwMutex.Unlock()
		}
	}
}

// Dump : list session info
func (fspgateway *FSPGateway) Dump() {
	for _, v := range fspgateway.mapSession {
		v.Info()
	}
	fspgateway.logger.Info("num of session :", len(fspgateway.mapSession))
}

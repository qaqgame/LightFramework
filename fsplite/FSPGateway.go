package fsplite

import (
	"github.com/golang/protobuf/proto"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// FSPGateway : FSP Gateway
type FSPGateway struct {
	mapSession map[uint32]*FSPSession
	IsRunning  bool
	Param      *FSPParam

	// udp conn related
	conn          *net.UDPConn
	receiveBuffer []byte
	port          int

	closeSignal          chan int
	lastClearSeesionTime int64

	recvRunning bool
	logger      *logrus.Entry
	rwMutex     sync.RWMutex
}

// NewSessionID : allocate a new Sid for session
func NewSessionID() uint32 {
	sid++
	return sid
}

// NewFSPGateway : create a fsplite gateway
func NewFSPGateway(_port int) *FSPGateway {
	fspgateway := new(FSPGateway)
	fspgateway.port = _port

	fspgateway.mapSession = make(map[uint32]*FSPSession)
	fspgateway.receiveBuffer = make([]byte, 4096)

	fspgateway.logger = logrus.WithFields(logrus.Fields{"Process": "FSPGateway"})
	fspgateway.IsRunning = false
	fspgateway.conn = nil
	fspgateway.closeSignal = make(chan int, 2)
	fspgateway.recvRunning = false
	fspgateway.lastClearSeesionTime = 0
	fspgateway.rwMutex = sync.RWMutex{}

	fspgateway.Start()
	return fspgateway
}

// CreateSession : create a fspliste session
func (fspgateway *FSPGateway) CreateSession() *FSPSession {
	sid := NewSessionID()
	// conv pharse is equal to sid of fsplite session for each session
	session := NewFSPSession(sid, fspgateway.HandleSessionSend)

	fspgateway.rwMutex.Lock()
	fspgateway.mapSession[sid] = session
	fspgateway.rwMutex.Unlock()
	return session
}

// ReleaseSession : clear session data stored in gateway
func (fspgateway *FSPGateway) ReleaseSession(sid uint32) {
	session := fspgateway.mapSession[sid]
	if session != nil {
		fspgateway.rwMutex.Lock()
		delete(fspgateway.mapSession, sid)
		fspgateway.rwMutex.Unlock()
	}
}

// HandleSessionSend : send data to remote client
func (fspgateway *FSPGateway) HandleSessionSend(endpoint *net.UDPAddr, bytes []byte, lenght int) {
	// fspgateway.logger.Debug("HandSessionSend in FSPGateway")
	if fspgateway.conn != nil {
		_, err := fspgateway.conn.WriteToUDP(bytes[:lenght], endpoint)
		// fspgateway.logger.Debug("HandleSessionSend In FSPGateway, wirte len: ", n)
		if err != nil {
			fspgateway.logger.Error("HandleSessionSend In FSPGateway, write failed")
		}
	} else {
		fspgateway.logger.Warn("HandleSessionSend in FSPGateway, conn is nil")
	}
}

// Clean : clear session data and close gateway.
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

	// listen specified prot
	udpAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:"+strconv.Itoa(fspgateway.port))
	if err != nil {
		fspgateway.logger.Fatal("Start Gateway error(resolveUDPAddr): ", err)
	}
	fspgateway.conn, err = net.ListenUDP("udp4", udpAddr)
	if err != nil {
		fspgateway.logger.Fatal("Start gateway err(listenUDP): ", err)
	}
	fspgateway.logger.Info("Listen UDP: ", fspgateway.port, udpAddr)

	// make sure all goroutine is running
	for _, v := range fspgateway.mapSession {
		if !v.ReceiveInMainActive {
			v.FSPSessionInit()
		}
	}

	// do receive data from remote client in gateway in another goroutine
	go fspgateway.Recv()
}

// Close : close gatewawy
func (fspgateway *FSPGateway) Close() {
	fspgateway.IsRunning = false

	for _, v := range fspgateway.mapSession {
		if v.ReceiveInMainActive {
			v.StopReceive()
		}
	}

	// if fspgateway.recvRunning {
	// 	fspgateway.closeSignal <- -1
	// }

	_ = fspgateway.conn.Close()
	fspgateway.conn = nil
}

// Recv : receive goroutine
func (fspgateway *FSPGateway) Recv() {
	fspgateway.logger.Debug("Start Recv Goroutine Of FSPGateway")
	fspgateway.recvRunning = true // a flag

	// receive data in a loop.
	for fspgateway.IsRunning {
		fspgateway.DoReceiveInGoroutine()
	}

	fspgateway.recvRunning = false
}

// GetSession : get Session from Gatway
func (fspgateway *FSPGateway) GetSession(sid uint32) *FSPSession {
	return fspgateway.mapSession[sid]
}

// DoReceiveInGoroutine : concrete operation of receive data from client
func (fspgateway *FSPGateway) DoReceiveInGoroutine() {
	// fspgateway.logger.Debug("In function DoReceiveGoroutine of FSPGateway")
	// sidbuf := make([]byte, 4)
	n, addr, err := fspgateway.conn.ReadFromUDP(fspgateway.receiveBuffer)
	if err != nil {
		fspgateway.logger.Error("error DoReceiveInGoroutine of FSPGateway: ", err)
	}
	// fspgateway.logger.Debug("Received data from UDP in FSPGateway, length is ", n)
	// data's lenght > 0

	if n <= 24 {
		return
	}

	if n > 0 {
		// use first four bits as sid
		// TODO: read here error:
		// sidbuf = fspgateway.receiveBuffer[24:28]
		tmp := fspgateway.receiveBuffer[24:n]
		// fspgateway.logger.Warn("Datas from conn: ", fspgateway.receiveBuffer[:n])

		fspmsg := new(FSPDataC2S)
		fspmsg.Msgs = make([]*FSPMessage,0)
		err := proto.Unmarshal(tmp, fspmsg)
		if err != nil {
			// fspgateway.logger.Error("Unmarshal msg error: ",err)
			// return
		}
		// fspgateway.logger.Info("all: ", fspgateway.receiveBuffer[:n])
		// fspgateway.logger.Info("tmp: ", tmp, "len: ", len(tmp))
		if fspmsg.String() != "" {
			fspgateway.logger.Warn("fspmsg: ",fspmsg)
		}
		var session *FSPSession = nil
		// tmp1 := binary.LittleEndian.Uint32(sidbuf)
		// fspgateway.logger.Warn("LittleEndian res: ", tmp1)
		sid := fspmsg.Sid
		if sid == 0 {
			// fspgateway.logger.Warn("Sid 为0，丢弃该包")
			return
		} else {
			fspgateway.logger.Warn("Sid 为 : ", sid)
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

// Tick : time signal,
// clear unactive session and tick active session to update their kcp
func (fspgateway *FSPGateway) Tick() {
	if fspgateway.IsRunning {
		discrepancy := uint32(time.Now().Sub(reftime) / time.Millisecond)
		currentTime := time.Now().Unix()
		if currentTime-fspgateway.lastClearSeesionTime > SessionActiveTimeout {
			fspgateway.lastClearSeesionTime = currentTime
			// clear session after a specified interval
			fspgateway.ClearNoActiveSession()
		}

		for _, v := range fspgateway.mapSession {
			// tick each session, use time as param
			v.Tick(discrepancy)
		}
	}
}

// ClearNoActiveSession : clear sessions which are no longer active
func (fspgateway *FSPGateway) ClearNoActiveSession() {
	for k, v := range fspgateway.mapSession {
		// clear
		if !v.IsActive() {
			fspgateway.rwMutex.Lock()
			fspgateway.logger.Fatal("session not active: ", v.sid)
			panic(v.sid)
			delete(fspgateway.mapSession, k)
			fspgateway.rwMutex.Unlock()
		}
	}
}

// Dump : list session info. for debugging using
func (fspgateway *FSPGateway) Dump() {
	for _, v := range fspgateway.mapSession {
		v.Info()
	}
	fspgateway.logger.Info("num of session :", len(fspgateway.mapSession))
}

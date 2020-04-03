package Server

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

type sessionId struct {
	lastId uint32
}

// ISession : reference a session between server and client
type ISession interface {
	GetUid() uint32
	GetId() uint32
	Ping() uint32
	SetPing(ping uint32)
	IsActive() bool
	IsAuth() bool
	SetAuth(userId uint32)
	Send(cnt []byte, length int) bool
	GetRemoteEndPoint() *net.UDPAddr
	Tick(currentTime uint32)
	DoReceiveInGateWay(buf []byte, size int)
	Active(addr *net.UDPAddr)
}

type ISessionListener interface {
	OnReceive(session ISession, bytes []byte, length int)
	GetLogger() *log.Entry
}

//var SId = &sessionId{0}
//var refTime time.Time = time.Now()

// SId : sessionid
var SId *sessionId
var refTime time.Time

func init() {
	log.WithFields(log.Fields{}).Info("Init ISession initialize data")
	SId = &sessionId{0}
	refTime = time.Now()
}

func (si *sessionId) NewId() uint32 {
	si.lastId++
	return si.lastId
}

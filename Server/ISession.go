package Server

import (
	"log"
	"net"
	"time"
)

type sessionId struct {
	lastId uint32
}

type ISession interface {
	GetUid() uint32
	GetId() uint32
	Ping() uint32
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
}

var SId = &sessionId{0}
var refTime time.Time = time.Now()


func Init() {
	log.Println("init sessionid")
	SId = &sessionId{0}
	refTime = time.Now()
}

func (si *sessionId) NewId() uint32 {
	si.lastId++
	return si.lastId
}


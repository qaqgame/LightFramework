package Server

import (
	"github.com/xtaci/kcp-go"
	"log"
	"net"
	"time"
)

var ActiveTimeout int64 = 30

type Sender func(session ISession, bytes []byte, length int)

type KCPSession struct {
	sid            uint32
	userid         uint32
	listener       ISessionListener
	sender         Sender
	remoteEndPoint *net.UDPAddr
	lastActiveTime int64
	nextUpdateTime uint32
	active         bool
	recvData       chan []byte
	needKCPUpdate  bool

	Kcp      *kcp.KCP
}

func NewKCPSession(_sid uint32, _sender Sender, _listener ISessionListener) *KCPSession {
	kcpSession := new(KCPSession)
	kcpSession.sid = _sid
	kcpSession.sender = _sender
	kcpSession.listener = _listener

	kcpSession.lastActiveTime = 0
	kcpSession.nextUpdateTime= 0
	kcpSession.active = false
	kcpSession.recvData = make(chan []byte,10)
	kcpSession.needKCPUpdate = false

	kcpSession.Kcp = kcp.NewKCP(_sid, kcpSession.HandKcpSend)
	kcpSession.Kcp.NoDelay(1,10,2,1)
	kcpSession.Kcp.WndSize(128,128)
	kcpSession.Initialize()
	return kcpSession
}

func (kcpSession *KCPSession) Initialize() {
	go kcpSession.DoReceiveInMain()
}

func (kcpSession *KCPSession) HandKcpSend(buf []byte, size int) {
	kcpSession.sender(kcpSession,buf,size)
}

func (kcpSession *KCPSession) Active(addr *net.UDPAddr) {
	kcpSession.lastActiveTime = time.Now().Unix()
	kcpSession.active = true

	kcpSession.remoteEndPoint = addr
}

func (kcpSession *KCPSession) GetUid() uint32 {

	return kcpSession.userid
}

func (kcpSession *KCPSession) GetId() uint32 {

	return kcpSession.sid
}

func (kcpSession *KCPSession) Ping() uint32 {

	return 0
}

func (kcpSession *KCPSession) IsActive() bool {
	if !kcpSession.active {
		return false
	} else {
		dt := time.Now().Unix() - kcpSession.lastActiveTime
		if dt > ActiveTimeout {
			kcpSession.active = false
		}
		return kcpSession.active
	}
}

func (kcpSession *KCPSession) IsAuth() bool {

	return kcpSession.userid > 0
}

func (kcpSession *KCPSession) SetAuth(userId uint32) {
	kcpSession.userid = userId
	return
}

func (kcpSession *KCPSession) Send(cnt []byte, length int) bool {
	if !kcpSession.IsActive() {
		log.Println("Client close")
		return false
	}
	return kcpSession.Kcp.Send(cnt) > 0
}

func (kcpSession *KCPSession) GetRemoteEndPoint() *net.UDPAddr {
	return kcpSession.remoteEndPoint
}

func (kcpSession *KCPSession) DoReceiveInGateWay(buf []byte, size int) {
	kcpSession.recvData <- buf
}

func (kcpSession *KCPSession) DoReceiveInMain() {
	for true {
		select {
		case data := <-kcpSession.recvData:
			ret := kcpSession.Kcp.Input(data,true,true)
			if ret < 0 {
				log.Println("not a correct package")
				return
			}
			kcpSession.needKCPUpdate = true
			for size := kcpSession.Kcp.PeekSize(); size > 0; size = kcpSession.Kcp.PeekSize()  {
				buf := make([]byte,size)
				if kcpSession.Kcp.Recv(buf) > 0 {
					kcpSession.listener.OnReceive(kcpSession, buf, size)
				}
			}
		}
	}
}

func (kcpSession *KCPSession) Tick(currentTime uint32) {
	kcpSession.DoReceiveInMain()
	current := currentTime
	if kcpSession.needKCPUpdate || current >= kcpSession.nextUpdateTime {
		kcpSession.Kcp.Update()
		kcpSession.nextUpdateTime = kcpSession.Kcp.Check()
		kcpSession.needKCPUpdate = false
	}
}
package Server

import (
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

type Gateway struct {
	mapSession      map[uint32]ISession
	conn            *net.UDPConn

	isRunning       bool
	recvBuf         []byte
	listener        ISessionListener
	port            int
	closeSignal     chan int
	lastClearSessionTime int64

	recvRunning     bool
	rwMutex         sync.RWMutex
}

func NewGateway(_port int, _listener ISessionListener) *Gateway {
	gateway := new(Gateway)
	gateway.port = _port
	gateway.listener = _listener

	gateway.mapSession = make(map[uint32]ISession)
	gateway.recvRunning = false
	gateway.isRunning = false
	gateway.conn = nil
	gateway.recvBuf = make([]byte,4096)
	gateway.closeSignal = make(chan int, 2)
	gateway.rwMutex = sync.RWMutex{}

	return gateway
}

func (gateway *Gateway) Init() {
	gateway.start()
}

func (gateway *Gateway) start() {
	gateway.isRunning = true

	udpAddr,err := net.ResolveUDPAddr("udp4","127.0.0.1:"+strconv.Itoa(gateway.port))
	if err != nil {
		log.Fatal("Start gateway err(resolveUDPAddr): ",err)
	}

	gateway.conn,err  = net.ListenUDP("udp",udpAddr)
	if err != nil {
		log.Fatal("Start gateway err(listenUdp): ", err)
	}

	go gateway.Recv()
}

func (gateway *Gateway) Clean() {
	gateway.mapSession = make(map[uint32]ISession)
	gateway.Close()
}

func (gateway *Gateway) Close() {
	gateway.isRunning = false

	if gateway.recvRunning {
		gateway.closeSignal <- -1
	}

	_ = gateway.conn.Close()
	gateway.conn = nil
}

func (gateway *Gateway) GetSession(sid uint32) ISession {
	return gateway.mapSession[sid]
}

// 接受数据的协程
func (gateway *Gateway) Recv() {
	gateway.recvRunning = true

	for gateway.isRunning {
		select {
		case v := <-gateway.closeSignal:
			if v == -1 {
				gateway.recvRunning = false
				return
			}
		default:
			// todo
			gateway.DoReceiveInGoroutine()
		}
	}

	gateway.recvRunning = false
}

func (gateway *Gateway) DoReceiveInGoroutine() {
	sidBuf := make([]byte,4)
	n, addr, err := gateway.conn.ReadFromUDP(gateway.recvBuf)
	if err != nil {
		log.Println("error DoReceiveInThread err: ",err)
	}
	if n > 0 {
		//buferReader := bytes.NewBuffer(gateway.recvBuf)
		//_,_ = buferReader.Read(sidBuf)

		sidBuf = gateway.recvBuf[:4]

		var kcpsession ISession = nil
		convId := binary.BigEndian.Uint32(sidBuf)
		if convId == 0 {
			sid := SId.NewId()
			kcpsession = NewKCPSession(sid, gateway.HandSessionSender, gateway.listener)
			gateway.rwMutex.Lock()
			gateway.mapSession[sid]=kcpsession
			gateway.rwMutex.Unlock()
		} else {
			kcpsession = gateway.mapSession[convId]
		}

		if kcpsession != nil {
			kcpsession.Active(addr)
			kcpsession.DoReceiveInGateWay(gateway.recvBuf,n)
		} else {
			log.Println("useless package in DoReceiveInGoroutine")
		}
	}
}

func (gateway *Gateway) HandSessionSender(session ISession,buf []byte, size int) {
	if gateway.conn != nil {
		_, err := gateway.conn.WriteToUDP(buf, session.GetRemoteEndPoint())
		if err != nil {
			log.Println("HandSessionSender error: ",err)
		}
	} else {
		log.Println("HandSessionSender: conn has been closed")
	}
}

func (gateway *Gateway) Tick() {
	if gateway.isRunning {
		discrepancy := uint32(time.Now().Sub(refTime) / time.Millisecond)
		currentTime := time.Now().Unix()
		if currentTime - gateway.lastClearSessionTime > ActiveTimeout {
			gateway.lastClearSessionTime = currentTime
			gateway.ClearNoActionSession()
		}

		for _,v := range gateway.mapSession {
			v.Tick(discrepancy)
		}
	}
}

func (gateway *Gateway) ClearNoActionSession() {
	for k,v := range gateway.mapSession {
		if !v.IsActive() {
			gateway.rwMutex.Lock()
			delete(gateway.mapSession,k)
			gateway.rwMutex.Unlock()
		}
	}
}
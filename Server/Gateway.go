package Server

import (
	"encoding/binary"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
}

// Gateway : struct of Gateway in Server
type Gateway struct {
	mapSession map[uint32]ISession
	conn       *net.UDPConn

	isRunning            bool
	recvBuf              []byte
	listener             ISessionListener
	port                 int
	closeSignal          chan int
	lastClearSessionTime int64

	recvRunning bool
	rwMutex     sync.RWMutex
	logger      *log.Entry
}

// NewGateway : new a Gatway via this function
func NewGateway(_port int, _listener ISessionListener) *Gateway {
	//log.Println("New Gateway")
	gateway := new(Gateway)
	gateway.port = _port
	gateway.listener = _listener
	gateway.logger = _listener.GetLogger()
	gateway.logger.Info("New a Gateway")

	gateway.mapSession = make(map[uint32]ISession)
	gateway.recvRunning = false
	gateway.isRunning = false
	gateway.conn = nil
	gateway.recvBuf = make([]byte, 4096)
	gateway.closeSignal = make(chan int, 2)
	gateway.rwMutex = sync.RWMutex{}

	gateway.Initialize()

	gateway.logger.Info("Gateway Created")
	return gateway
}

// Initialize : start a Gatway
func (gateway *Gateway) Initialize() {
	gateway.start()
}

func (gateway *Gateway) start() {
	gateway.logger.Info("Gateway Started")
	gateway.isRunning = true

	udpAddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:"+strconv.Itoa(gateway.port))
	if err != nil {
		log.Fatal("Start gateway err(resolveUDPAddr): ", err)
	}

	gateway.conn, err = net.ListenUDP("udp4", udpAddr)
	if err != nil {
		log.Fatal("Start gateway err(listenUdp): ", err)
	}
	//
	//log.Println("Listen udp:",gateway.port,udpAddr)
	gateway.logger.Info("Listen UDP: ", gateway.port, udpAddr)

	go gateway.Recv()
}

// Clean : clean the Gatway's storage
func (gateway *Gateway) Clean() {
	gateway.mapSession = make(map[uint32]ISession)
	gateway.Close()
}

// Close : close Gatway
func (gateway *Gateway) Close() {
	gateway.isRunning = false

	for _, v := range gateway.mapSession {
		v.StopReceive()
	}
	// if gateway.recvRunning {
	// 	gateway.closeSignal <- -1
	// }

	_ = gateway.conn.Close()
	gateway.conn = nil
}

// GetSession : get Session from Gatway
func (gateway *Gateway) GetSession(sid uint32) ISession {
	return gateway.mapSession[sid]
}

// Recv : 接受数据的协程
func (gateway *Gateway) Recv() {
	gateway.logger.Info("Start Recv Goroutine of Gateway")
	gateway.recvRunning = true

	for gateway.isRunning {
		gateway.DoReceiveInGoroutine()
	}

	gateway.recvRunning = false
}

// DoReceiveInGoroutine : recive infos form udp connection in another goroutine
func (gateway *Gateway) DoReceiveInGoroutine() {
	// gateway.logger.Info("in function DoReceiveInGoroutine of Gateway")
	sidBuf := make([]byte, 4)

	// lis,err := kcp.DialWithOptions(":"+strconv.Itoa(gateway.port),nil,0,0)

	n, addr, err := gateway.conn.ReadFromUDP(gateway.recvBuf)
	// gateway.logger.Warn("remote addr 1: ", addr, "info: ", gateway.recvBuf[:n])
	if err != nil {
		//log.Println("error DoReceiveInThread err: ",err)
		gateway.logger.Error("error DoReceiveInGoroutine of Gateway: ", err)
	}
	//log.Println(time.Now().Unix(),"Received data: ", n)
	// gateway.logger.Debug("Received data from UDP in Gateway, length is ", n)
	if n >= 44 {
		sidBuf = gateway.recvBuf[24:28]

		var kcpsession ISession = nil
		uid := binary.BigEndian.Uint32(sidBuf)
		//log.Println("read ",uid)
		// gateway.logger.Debug("Uid part of ProtocolHead is ", uid," sidbuf: ", sidBuf)
		if uid == 0 {

			for k, v := range gateway.mapSession {
				if v.GetRemoteEndPoint().String() == addr.String() {
					kcpsession = gateway.mapSession[k]
					break
				}
			}
			if kcpsession == nil {
				// gateway.logger.Debug("Uid is 0")
				sid := SId.NewId()
				kcpsession = NewKCPSession(sid, gateway.HandSessionSender, gateway.listener, 1, gateway.logger)
				// gateway.logger.Debug("KCPSession created is : ", sid)
				gateway.rwMutex.Lock()
				gateway.mapSession[sid] = kcpsession
				gateway.rwMutex.Unlock()
			} else {
				kcpsession = nil
			}
		} else {
			// gateway.logger.Debug("Uid isn't 0 but is ", uid)
			kcpsession = gateway.mapSession[uid]
		}

		if kcpsession != nil {
			// gateway.logger.Warn("info: ",gateway.recvBuf[:n] )
			kcpsession.Active(addr)
			kcpsession.DoReceiveInGateWay(gateway.recvBuf, n)
		} else {
			gateway.logger.Warn("DeReceiveInGoroutine of Gateway, KCPSession is nil")
		}
	}
}

// HandSessionSender : callback function of kcp
func (gateway *Gateway) HandSessionSender(session ISession, buf []byte, size int) {
	// gateway.logger.Info("HandSessionSender in Gateway, session's id is ", session.GetId())
	// gateway.logger.Info("HandSessionSender in Gateway, data size is ", size, len(buf))

	// gateway.logger.Debug("HandSessionSender in Gateway, data is: ", buf[:size])

	if gateway.conn != nil {
		_, err := gateway.conn.WriteToUDP(buf[:size], session.GetRemoteEndPoint())
		//log.Println("写了",n,"字节")
		// gateway.logger.Info("HandSessionSender in Gateway, Write", n, "byte to UDPConn")
		if err != nil {
			gateway.logger.Error("HandSessionSender in Gateway, Write to UDPConn error", err)
		}
	} else {
		gateway.logger.Warn("HandSessionSender in Gateway, UDPConn has been closed")
	}

	// gateway.logger.Info("HandSessionSender in Gateway, End of write to UDPConn")
}

// Tick : tick Gatway
func (gateway *Gateway) Tick() {
	if gateway.isRunning {
		discrepancy := uint32(time.Now().Sub(refTime) / time.Millisecond)
		currentTime := time.Now().Unix()
		if currentTime-gateway.lastClearSessionTime > ActiveTimeout {
			gateway.lastClearSessionTime = currentTime
			gateway.ClearNoActionSession()
		}

		for _, v := range gateway.mapSession {
			v.Tick(discrepancy)
		}
	}
}

// ClearNoActionSession : clear storage of Gatway which session is not active any more
func (gateway *Gateway) ClearNoActionSession() {
	for k, v := range gateway.mapSession {
		if !v.IsActive() {
			gateway.rwMutex.Lock()
			delete(gateway.mapSession, k)
			gateway.rwMutex.Unlock()
		}
	}
}

// Dump : list session info
func (gateway *Gateway) Dump() {
	for _, v := range gateway.mapSession {
		gateway.logger.Info("session id:", v.GetId(), "session uid: ", v.GetUid(), "is active :", v.IsActive())
	}
	gateway.logger.Info("num of session :", len(gateway.mapSession))
}

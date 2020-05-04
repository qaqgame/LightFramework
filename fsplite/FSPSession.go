package fsplite

import (
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go"
)

// FSPSession :
type FSPSession struct {
	sid               uint32 // session id
	ping              uint32
	sender            FSPSender
	listener          FSPListener
	remoteEndPoint    *net.UDPAddr
	isEndPointChanged bool // remote client notwork environment change

	// KCP related
	Kcp               *kcp.KCP
	recvData          chan []byte
	nextkcpupdatetime uint32
	needkcpupdateflag bool

	closeDoreceive chan int
	sendBufData    *FSPDataS2C

	lastActiveTime      int64
	logger              *logrus.Entry
	isActive            bool
	ReceiveInMainActive bool // a flag showing if the goroutine is running
}

// FSPSender : deliver data to upper level, and the upper level will do the send operation.
type FSPSender func(endpoint *net.UDPAddr, bytes []byte, lenght int)

// FSPListener : data listener
type FSPListener func(msg *FSPDataC2S)

// NewFSPSession : create a fsplite session
func NewFSPSession(_sid uint32, _sender FSPSender) *FSPSession {

	fspsesssion := new(FSPSession)

	fspsesssion.sid = _sid
	fspsesssion.sender = _sender

	// create KCP part
	fspsesssion.Kcp = kcp.NewKCP(2, fspsesssion.HandKcpSend)
	fspsesssion.Kcp.NoDelay(1, 20, 2, 1)
	fspsesssion.Kcp.WndSize(128, 128)

	fspsesssion.remoteEndPoint = nil
	fspsesssion.recvData = make(chan []byte, 128)
	fspsesssion.closeDoreceive = make(chan int, 2)
	fspsesssion.needkcpupdateflag = false
	fspsesssion.nextkcpupdatetime = 0
	fspsesssion.isEndPointChanged = false

	fspsesssion.logger = logrus.WithFields(logrus.Fields{"Process": "FSPSession"})

	fspsesssion.sendBufData = new(FSPDataS2C)

	fspsesssion.FSPSessionInit()
	fspsesssion.logger.Debug("New FSPSession")
	return fspsesssion
}

// GetSid : return sid
func (fspsession *FSPSession) GetSid() uint32 {
	return fspsession.sid
}

// GetPing : return ping
func (fspsession *FSPSession) GetPing() uint32 {
	return fspsession.ping
}

// HandKcpSend : callback handler of Kcp
func (fspsession *FSPSession) HandKcpSend(buffer []byte, size int) {
	fspsession.sender(fspsession.remoteEndPoint, buffer, size)
}

// SetReceiveListener : set listener of fspsession
func (fspsession *FSPSession) SetReceiveListener(_listener FSPListener) {
	fspsession.listener = _listener
}

// FSPSessionInit : start a new goroutine to run handling data.
func (fspsession *FSPSession) FSPSessionInit() {
	go fspsession.DoReceiveInMain()
}

// Active : active session
func (fspsession *FSPSession) Active(addr *net.UDPAddr) {
	fspsession.lastActiveTime = time.Now().Unix()
	fspsession.isActive = true

	// remote client network circumstance changes
	if fspsession.remoteEndPoint == nil || fspsession.remoteEndPoint.String() != addr.String() {
		fspsession.isEndPointChanged = true
		fspsession.remoteEndPoint = addr
	}
}

// SetAuth : set auth
func (fspsession *FSPSession) SetAuth() {
	fspsession.isEndPointChanged = false
}

// IsActive : return active state
func (fspsession *FSPSession) IsActive() bool {
	if !fspsession.isActive {
		return false
	}

	des := time.Now().Unix() - fspsession.lastActiveTime
	if des > SessionActiveTimeout {
		fspsession.isActive = false
	}
	return fspsession.isActive
}

// Send : send data
func (fspsession *FSPSession) Send(frame *FSPFrame) bool {
	if !fspsession.IsActive() {
		// fspsession.logger.Warn("FSPSession is not active", fspsession.sid)
		return false
	}
	fspsession.sendBufData.Frames = make([]*FSPFrame, 0)
	fspsession.sendBufData.Frames = append(fspsession.sendBufData.Frames, frame)
	data, err := proto.Marshal(fspsession.sendBufData)
	if err != nil {
		fspsession.logger.Warn("Marshal FSPDataS2C error : ", err)
		return false
	}

	i := fspsession.Kcp.Send(data)
	return i == 0
}

// StopReceive : signal of close goroutine : DoReceiveInMain
func (fspsession *FSPSession) StopReceive() {
	fspsession.closeDoreceive <- -1
}

// DoReceiveInGateway : after receive data in function DoReceiveInGateway of fsplite.gateway,
// data will be sent to channel recvData for corresponding session
func (fspsession *FSPSession) DoReceiveInGateway(buf []byte, size int) {
	fspsession.recvData <- buf[:size]
}

// DoReceiveInMain : a independent goroutine.
// keep running as long as session created, read data from channel recvData and handle it.
func (fspsession *FSPSession) DoReceiveInMain() {
	// a flag to show if the goroutine is running.
	fspsession.ReceiveInMainActive = true
	for true {
		select {
		// Reading from recvdata channal.
		case data := <-fspsession.recvData:
			fspsession.logger.Debug("DoReceiveInMain of FSPSession received data len: ", len(data))
			ret := fspsession.Kcp.Input(data, true, true)
			if ret < 0 {
				fspsession.logger.Warn("DeReceiveInMain of FSPSession, data is not a correct pakcage, input ret : ", ret)
				continue
			}
			fspsession.needkcpupdateflag = true
			for size := fspsession.Kcp.PeekSize(); size > 0; size = fspsession.Kcp.PeekSize() {
				buf := make([]byte, size)
				if fspsession.Kcp.Recv(buf) > 0 {
					if fspsession.listener != nil {
						data := new(FSPDataC2S)
						err := proto.Unmarshal(buf, data)
						if err != nil {
							fspsession.logger.Warn("Can not Unmarsh as a proto message")
							continue
						}
						// handle data received
						fspsession.listener(data)
					} else {
						fspsession.logger.Warn("找不到接收者")
					}
				}
			}
		// channel to control the goroutine
		case _ = <-fspsession.closeDoreceive:
			// set to false
			fspsession.ReceiveInMainActive = false
			return
		}
	}
}

// Tick : tick session, update kcp state
func (fspsession *FSPSession) Tick(currentTime uint32) {
	current := currentTime
	if fspsession.needkcpupdateflag || current >= fspsession.nextkcpupdatetime {
		fspsession.Kcp.Update()
		fspsession.nextkcpupdatetime = fspsession.Kcp.Check()
		fspsession.needkcpupdateflag = false
	}
}

// Info : debugging using
func (fspsession *FSPSession) Info() {
	fspsession.logger.Info("fspsession id: ", fspsession.sid, "fspsession isactive:", fspsession.isActive)
}

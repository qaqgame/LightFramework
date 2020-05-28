package fsplite

import (
	"fmt"
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
	// datachan1         *chan []byte
	// datachan2         *chan []byte
	//---------
	dataQueue1        *Queue
	dataQueue2        *Queue
	//----------
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
	fspsesssion.Kcp.NoDelay(1, 10, 2, 1)
	fspsesssion.Kcp.WndSize(128, 128)

	fspsesssion.remoteEndPoint = nil
	fspsesssion.recvData = make(chan []byte, 128)
	//tmp1 := make(chan []byte, 128)
	//fspsesssion.datachan1 = &tmp1
	//tmp2 := make(chan []byte, 128)
	//fspsesssion.datachan2 = &tmp2
	//---------
	fspsesssion.dataQueue1 = NewQueue()
	fspsesssion.dataQueue2 = NewQueue()
	//---------

	fspsesssion.closeDoreceive = make(chan int, 2)
	fspsesssion.needkcpupdateflag = false
	fspsesssion.nextkcpupdatetime = 0
	fspsesssion.isEndPointChanged = false

	fspsesssion.logger = logrus.WithFields(logrus.Fields{"Process": "FSPSession"})

	fspsesssion.sendBufData = new(FSPDataS2C)

	// fspsesssion.FSPSessionInit()
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
	if fspsession.remoteEndPoint == nil || fspsession.remoteEndPoint.IP.String() != addr.IP.String() {
		logrus.Warn("RemoteEndPoint Changed", addr.String()," ", fspsession.remoteEndPoint.String())
		fspsession.isEndPointChanged = true
		// fspsession.isEndPointChanged = true
	}
	fspsession.remoteEndPoint = addr
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
	if des > SessionActiveTimeout{
		logrus.Warn("des: ",des)
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
	return i >= 0
}

// StopReceive : signal of close goroutine : DoReceiveInMain
//func (fspsession *FSPSession) StopReceive() {
//	fspsession.closeDoreceive <- -1
//}

// DoReceiveInGateway : after receive data in function DoReceiveInGateway of fsplite.gateway,
// data will be sent to channel recvData for corresponding session
func (fspsession *FSPSession) DoReceiveInGateway(buf []byte, size int) {
	// fspsession.recvData <- buf[:size]
	v := buf[:size]
	// *fspsession.datachan1 <- v
	//----------
	fspsession.dataQueue1.Push(v)
	//----------
}

// DoReceiveInMain : a independent goroutine.
// keep running as long as session created, read data from channel recvData and handle it.
func (fspsession *FSPSession) DoReceiveInMain() {
	// a flag to show if the goroutine is running.

	//-----Origin
	//fspsession.ReceiveInMainActive = true
	//fspsession.datachan1, fspsession.datachan2 = fspsession.datachan2, fspsession.datachan1
	//
	//for true {
	//	select {
	//	// Reading from recvdata channal.
	//	case data := <- *fspsession.datachan2:
	//		// fspsession.logger.Debug("DoReceiveInMain of FSPSession received data len: ", len(data))
	//		ret := fspsession.Kcp.Input(data, true, true)
	//		if ret < 0 {
	//			fspsession.logger.Warn("DeReceiveInMain of FSPSession, data is not a correct pakcage, input ret : ", ret)
	//			continue
	//		}
	//		fspsession.needkcpupdateflag = true
	//		for size := fspsession.Kcp.PeekSize(); size > 0; size = fspsession.Kcp.PeekSize() {
	//			buf := make([]byte, size)
	//			if fspsession.Kcp.Recv(buf) > 0 {
	//				if fspsession.listener != nil {
	//					data := new(FSPDataC2S)
	//					err := proto.Unmarshal(buf, data)
	//					if err != nil {
	//						fspsession.logger.Warn("Can not Unmarsh as a proto message")
	//						continue
	//					}
	//					// handle data received
	//					fspsession.listener(data)
	//				} else {
	//					fspsession.logger.Warn("找不到接收者")
	//				}
	//			}
	//		}
	//	default:
	//		return
	//	}
	//}
	//----

	//-----New
	fspsession.dataQueue1,fspsession.dataQueue2 = fspsession.dataQueue2,fspsession.dataQueue1
	for fspsession.dataQueue2.Len() > 0 {
		data := fspsession.dataQueue2.Pop().([]byte)
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
	}
	//-----
}

func (fspsession *FSPSession) HeartBeat() {
	fspheartbeat := FSPDataS2C{
		Frames: []*FSPFrame{
			{FrameID:0, Msgs: []*FSPMessage{{Cmd:10}}},
		},
	}

	v,err := proto.Marshal(&fspheartbeat)
	if err != nil {
		fmt.Println("Heart Beat to client error: ",err)
	}
	fspsession.Kcp.Send(v)
}

// Tick : tick session, update kcp state
func (fspsession *FSPSession) Tick(currentTime uint32) {
	// fspsession.HeartBeat()
	fspsession.DoReceiveInMain()

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

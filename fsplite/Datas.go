package fsplite

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"time"
	"errors"
)

const (
	// ServerFrameInterval : ms of frames interval
	ServerFrameInterval = 66
	// ServerTimeout : ms of client timeout
	ServerTimeout = 15000
	// ClientFrameRateMultiple :
	ClientFrameRateMultiple = 2
	// MaxFrameID :
	MaxFrameID = -1
	// AUTH :
	AUTH = 1008
	// EnableSpeedUp :
	EnableSpeedUp = true
	// DefaultSpeedUp :
	DefaultSpeedUp = 1
	// JitterBufferSize :
	JitterBufferSize = 0
	// EnableAutoBuffer :
	EnableAutoBuffer = true
	// SessionActiveTimeout :
	SessionActiveTimeout = 30
)

// GameStates
const (
	// None : 初始状态
	None = iota + 1000
	// Create : 游戏创建状态
	Create
	// GameBegin : 游戏开始状态
	GameBegin
	// RoundBegin : 回合开始
	RoundBegin
	// ControlStart : 可以开始操作
	ControlStart
	// RoundEnd : 回合结束
	RoundEnd
	// GameEnd : 游戏结束
	GameEndMsg
	// GameExit :
	GameExit
	// GameEndMsg
	GameEnd = 8
)

const (
	// NormalExit :
	NormalExit = iota + 100
)

// Empty : judge if fspframe is empty
func (fspframe *FSPFrame) Empty() bool {
	return len(fspframe.Msgs) == 0 || fspframe.Msgs == nil
}

var reftime time.Time
var sid uint32 = 0
var randomSeed int32
func init() {
	reftime = time.Now()
	rand.Seed(time.Now().UnixNano())
	randomSeed = rand.Int31()
}

func externalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("connected to the network?")
}

func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}
	return ip
}

func get_external() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(resp.Body)
	//s := buf.String()
	return string(content)
}

// NewDefaultFspParam : Create a default FspPrame
func NewDefaultFspParam(host string, port int, defaultModel int) *FSPParam {
	fspparam := new(FSPParam)
	fmt.Println("input type",defaultModel)
	// default param
	var ip string
	if defaultModel==0 {
		ip = get_external()
	} else if defaultModel == 1 {
		ip1,err := externalIP()
		if err != nil {
			ip1 = nil
		} else {
			ip = ip1.String()
		}
	}

	fspparam.Host = ip
	fmt.Println(fspparam.Host)
	fspparam.Port = int32(port)
	fspparam.Sid = 0
	fspparam.ServerFrameInterval = ServerFrameInterval
	fspparam.ServerTimeout = ServerTimeout
	fspparam.ClientFrameRateMultiple = ClientFrameRateMultiple
	fspparam.UseLocal = false
	fspparam.AuthID = AUTH
	fspparam.MaxFrameID = MaxFrameID
	fspparam.EnableSpeedUp = EnableSpeedUp
	fspparam.EnableAutoBuffer = EnableAutoBuffer
	fspparam.DefaultSpeed = DefaultSpeedUp
	fspparam.JitterBufferSize = JitterBufferSize
	fspparam.RandomSeed = randomSeed

	return fspparam
}

// TODO: use interface
type GameProcess interface {

}

type FSPGameI interface {
	OnStateGameCreate()
	OnStateGameBegin()
	OnStateRoundBegin()
	OnStateControlStart()
	OnStateRoundEnd()
	OnStateGameEnd()
	IsGameEnd() bool
	SetGameState(int, int, int)
	EnterFrame()
	AddCmdToCurrFrame(int32, []byte)
	AddMsgToCurrFrame(uint32, *FSPMessage)
	Release()
	AddPlayer(uint32, *FSPSession, uint32, uint32, uint32) *FSPPlayer
	GetGameID() uint32
	// session中收到msg时调用
	OnGameBeginCallBack(*FSPPlayer, *FSPMessage)
	// 将消息添加到frame后调用
	OnGameBeginMsgAddCallBack()
	// 生成GameBegin消息
	CreateGameBeginMsg() []byte

	OnRoundBeginCallBack(*FSPPlayer, *FSPMessage)
	OnRoundBeginMsgAddCallBack()
	CreateRoundMsg() []byte

	OnControlStartCallBack(*FSPPlayer, *FSPMessage)
	OnControlStartMsgAddCallBack()
	CreateControlStartMsg() []byte

	OnRoundEndCallBack(*FSPPlayer, *FSPMessage)
	OnRoundEndMsgAddCallBack()
	CreateRoundEndMsg() []byte

	OnGameEndCallBack(*FSPPlayer, *FSPMessage)
	OnGameEndMsgAddCallBack()
	CreateGameEndMsg() []byte
}
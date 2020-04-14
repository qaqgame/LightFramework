package fsplite

import (
	"time"
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
	AUTH = 10
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
	GameEnd
	// GameExit :
	GameExit
)

// Empty :
func (fspframe *FSPFrame) Empty() bool {
	return len(fspframe.Msgs) == 0 || fspframe.Msgs == nil
}

var reftime time.Time
var sid uint32 = 0

func init() {
	reftime = time.Now()
}

// NewDefaultFspParam :
func NewDefaultFspParam() *FSPParam {
	fspparam := new(FSPParam)
	// TODO: default param
	return fspparam
}

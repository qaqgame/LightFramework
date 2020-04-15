package fsplite

import "time"

// FSPManager :
type FSPManager struct {
	gateway             *FSPGateway
	mapGame             map[uint32]*FSPGame
	param               *FSPParam
	lastticks           int64
	useCostomEnterFrame bool // 用户自定义方式
}

// NewFSPManager :
func NewFSPManager(_port int) *FSPManager {
	fspmanager := new(FSPManager)
	fspmanager.mapGame = make(map[uint32]*FSPGame)
	fspmanager.gateway = NewFSPGateway(_port)
	// TODO: use default param creator
	fspmanager.param = new(FSPParam)
	fspmanager.lastticks = 0
	fspmanager.useCostomEnterFrame = false

	return fspmanager
}

// Clean :
func (fspmanager *FSPManager) Clean() {
	fspmanager.mapGame = make(map[uint32]*FSPGame)
	fspmanager.gateway.Clean()
}

// SetFSPInterval : 设置服务端帧间隔，客户端与服务端帧率比
func (fspmanager *FSPManager) SetFSPInterval(_serverFrameInterval, _clientFrameRateMultiple int32) {
	fspmanager.param.ServerFrameInterval = _serverFrameInterval
	fspmanager.param.ClientFrameRateMultiple = _clientFrameRateMultiple
}

// SetServerTimeout : 设置超时时间
func (fspmanager *FSPManager) SetServerTimeout(_serverTimeout int32) {
	fspmanager.param.ServerTimeout = _serverTimeout
}

// GetFrameInterval :
func (fspmanager *FSPManager) GetFrameInterval() int32 {
	return fspmanager.param.ServerFrameInterval
}

// GetParam :
func (fspmanager *FSPManager) GetParam() *FSPParam {
	return fspmanager.param
}

// Tick :
func (fspmanager *FSPManager) Tick() {
	fspmanager.gateway.Tick()

	nowtimenano := time.Now().UnixNano()
	interval := nowtimenano - fspmanager.lastticks
	if interval > int64(fspmanager.param.ServerFrameInterval*1e6) {
		fspmanager.lastticks = nowtimenano - nowtimenano%int64(fspmanager.param.ServerFrameInterval)

		if !fspmanager.useCostomEnterFrame {
			fspmanager.EnterFrame()
		}
	}
}

// EnterFrame :
func (fspmanager *FSPManager) EnterFrame() {
	if fspmanager.gateway.IsRunning {
		for _, v := range fspmanager.mapGame {
			v.EnterFrame()
		}
	}
}

// CreateGame :
func (fspmanager *FSPManager) CreateGame(gameid uint32) *FSPGame {
	// todo -
	fspgame := NewFSPGame(gameid, fspmanager.param)

	fspmanager.mapGame[gameid] = fspgame
	return fspgame
}

// ReleaseGame :
func (fspmanager *FSPManager) ReleaseGame(gameid uint32) {
	game := fspmanager.mapGame[gameid]
	if game != nil {
		game.Release()
		delete(fspmanager.mapGame, gameid)
	}
}

// AddPlayer :
func (fspmanager *FSPManager) AddPlayer(gameid, playerid uint32) uint32 {
	game := fspmanager.mapGame[gameid]
	session := fspmanager.gateway.CreateSession()

	game.AddPlayer(playerid, session)
	return session.GetSid()
}

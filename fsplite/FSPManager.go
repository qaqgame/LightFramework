package fsplite

import (
	"io/ioutil"
	"net/http"
	"time"
)

// FSPManager : fsplite manager
type FSPManager struct {
	gateway             *FSPGateway
	mapGame             map[uint32]FSPGameI
	param               *FSPParam
	lastticks           int64
	useCostomEnterFrame bool // 用户自定义方式
}

func getExternalIP() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content)
}

// NewFSPManager : create a manager
func NewFSPManager(_port int) *FSPManager {
	fspmanager := new(FSPManager)
	fspmanager.mapGame = make(map[uint32]FSPGameI)
	// gateway will automatically start after created
	fspmanager.gateway = NewFSPGateway(_port)
	// use default param creator to crate a new param
	fspmanager.param = NewDefaultFspParam(getExternalIP(), _port)
	fspmanager.lastticks = 0
	fspmanager.useCostomEnterFrame = false
	return fspmanager
}

// Clean : clear data
func (fspmanager *FSPManager) Clean() {
	fspmanager.mapGame = make(map[uint32]FSPGameI)
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

// GetFrameInterval : return frame interval
func (fspmanager *FSPManager) GetFrameInterval() int32 {
	return fspmanager.param.ServerFrameInterval
}

// GetParam : return fspparam
func (fspmanager *FSPManager) GetParam() *FSPParam {
	return fspmanager.param
}

// Tick : time signal
func (fspmanager *FSPManager) Tick() {
	// tick gateway
	fspmanager.gateway.Tick()

	nowtimenano := time.Now().UnixNano()
	interval := nowtimenano - fspmanager.lastticks
	// after the interval, then tick game to enterframe(update game state and send fspmsg to client)
	if interval > int64(fspmanager.param.ServerFrameInterval*1e6) {
		fspmanager.lastticks = nowtimenano - nowtimenano%int64(fspmanager.param.ServerFrameInterval)

		// user don't support a new EnterFrame Function
		if !fspmanager.useCostomEnterFrame {
			// default enterframe
			fspmanager.EnterFrame()
		}

		// TODO: user can use their own enterframe
	}
}

// EnterFrame : nitify all games to EnterFrame
func (fspmanager *FSPManager) EnterFrame() {
	if fspmanager.gateway.IsRunning {
		for _, v := range fspmanager.mapGame {
			v.EnterFrame()
		}
	}
}

// CreateGameI : create a game fits interface
func (fspmanager *FSPManager) CreateGameI(gameid uint32) FSPGameI{
	fspgame := NewFSPGame(gameid, fspmanager.param)

	fspmanager.mapGame[gameid] = fspgame
	return fspgame
}

// AddUDefinedGame:
func (fspmanager *FSPManager) AddUDefinedGame(i FSPGameI) {
	fspmanager.mapGame[i.GetGameID()] = i
}

// ReleaseGame : relase a specified game
func (fspmanager *FSPManager) ReleaseGame(gameid uint32) {
	game := fspmanager.mapGame[gameid]
	if game != nil {
		game.Release()
		delete(fspmanager.mapGame, gameid)
	}
}

// AddPlayer : add a player to a specified game, use player's uid as param, and retrun session's sid
func (fspmanager *FSPManager) AddPlayer(gameid, playerid, idInGame uint32) uint32 {
	game := fspmanager.mapGame[gameid]
	session := fspmanager.gateway.CreateSession()

	game.AddPlayer(playerid, session, idInGame, 0,0)
	return session.GetSid()
}

// AddPlayers : add players to specified game. return a map which key is player uid and value is player session's sid
func (fspmanager *FSPManager) AddPlayers(gameid uint32, playerids, fridenmask, enemymask map[uint32]uint32) map[uint32]uint32 {
	game := fspmanager.mapGame[gameid]
	sessionids := make(map[uint32]uint32)

	//key: playerId   value: id in game
	for k, v := range playerids {
		session := fspmanager.gateway.CreateSession()
		game.AddPlayer(k, session, v, fridenmask[v], enemymask[v])
		sessionids[k] = session.GetSid()
	}

	return sessionids
}
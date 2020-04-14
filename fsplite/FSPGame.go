package fsplite

import (
	"github.com/sirupsen/logrus"
)

// FSPGame :
type FSPGame struct {
	gameID       uint32
	maxplayerNum int32 // 10
	State        int
	stateParam1  int
	stateParam2  int
	param        *FSPParam

	gameBeginFlag    int16
	roundBeginFlag   int16
	controlStartFlag int16
	roundEndFlg      int16
	gameEndFlag      int16

	onGameExit OnGameExit
	onGameEnd  OnGameEnd

	CurrRoundID int16

	LockedFrame *FSPFrame
	playerList  map[uint32]*FSPPlayer
	logger      *logrus.Entry
}

// OnGameExit :
type OnGameExit func(uint32)

// OnGameEnd :
type OnGameEnd func(int32)

// NewFSPGame :
func NewFSPGame(gameid uint32, fspparam *FSPParam) *FSPGame {
	fspgame := new(FSPGame)
	fspgame.gameID = gameid
	fspgame.maxplayerNum = 10
	fspgame.LockedFrame = new(FSPFrame)
	fspgame.CurrRoundID = 0
	fspgame.playerList = make(map[uint32]*FSPPlayer)
	fspgame.param = fspparam

	fspgame.logger = logrus.WithFields(logrus.Fields{"GameID": fspgame.gameID})

	fspgame.SetGameState(Create, 0, 0)
	return fspgame
}

// Release :
func (fspgame *FSPGame) Release() {
	fspgame.SetGameState(None, 0, 0)
	for _, v := range fspgame.playerList {
		v.Release()
	}
	fspgame.playerList = make(map[uint32]*FSPPlayer)
}

// AddPlayer :
func (fspgame *FSPGame) AddPlayer(playerid uint32, session *FSPSession) *FSPPlayer {
	if fspgame.State != Create {
		fspgame.logger.Warn("Unable to add player, current state is :", fspgame.State)
		return nil
	}

	if len(fspgame.playerList) >= int(fspgame.maxplayerNum) {
		fspgame.logger.Warn("player counts have reached the max value of this game. current player counts: ", len(fspgame.playerList))
		return nil
	}

	if fspgame.playerList[playerid] != nil {
		fspgame.logger.Warn("player already exist, use new info to replace old info")
	}

	player := NewFSPPlayer(playerid, session, fspgame.param.AuthID, fspgame.OnRecvFromPlayer)
	fspgame.playerList[playerid] = player
	return player
}

// GetPlayer :
func (fspgame *FSPGame) GetPlayer(playerid uint32) *FSPPlayer {
	return fspgame.playerList[playerid]
}

// GetPlayerMap :
func (fspgame *FSPGame) GetPlayerMap() map[uint32]*FSPPlayer {
	return fspgame.playerList
}

// OnRecvFromPlayer :
func (fspgame *FSPGame) OnRecvFromPlayer(player *FSPPlayer, msg *FSPMessage) {
	fspgame.handleClientCmd(player, msg)
}

// handleClientCmd :
func (fspgame *FSPGame) handleClientCmd(player *FSPPlayer, msg *FSPMessage) {
	playerID := player.ID
	if !player.HasAuthed() {
		if msg.Cmd == AUTH {
			player.SetAuth(msg.Cmd)
		}
		return
	}

	switch msg.Cmd {
	case GameBegin:
		fspgame.SetFlag(playerID, &fspgame.gameBeginFlag, "gamebeginflag")
		fspgame.logger.Debug("gamebeginflag value: ", fspgame.gameBeginFlag)
	case RoundBegin:
		fspgame.SetFlag(playerID, &fspgame.roundBeginFlag, "roundbeginflag")
		fspgame.logger.Debug("roundbeginflag value: ", fspgame.roundBeginFlag)
	case ControlStart:
		fspgame.SetFlag(playerID, &fspgame.controlStartFlag, "controlstartflag")
		fspgame.logger.Debug("controlstartflag value: ", fspgame.controlStartFlag)
	case RoundEnd:
		fspgame.SetFlag(playerID, &fspgame.gameEndFlag, "gameendflag")
		fspgame.logger.Debug("roundendflag value: ", fspgame.gameEndFlag)
	case GameEnd:
		fspgame.SetFlag(playerID, &fspgame.gameEndFlag, "gameendflag")
		fspgame.logger.Debug("gameendflag value: ", fspgame.gameEndFlag)
	case GameExit:
		fspgame.HandleGameExit(playerID, msg)
		fspgame.logger.Debug("receive gameexit cmd from player id :", playerID)
	default:
		fspgame.AddMsgToCurrFrame(playerID, msg)
	}
}

// HandleGameExit :
func (fspgame *FSPGame) HandleGameExit(playerID uint32, msg *FSPMessage) {
	fspgame.AddMsgToCurrFrame(playerID, msg)
	player := fspgame.GetPlayer(playerID)
	if player != nil {
		player.WaitForExit = true

		if fspgame.onGameExit != nil {
			fspgame.onGameExit(player.ID)
		}
	}
}

// AddMsgToCurrFrame :
func (fspgame *FSPGame) AddMsgToCurrFrame(playerID uint32, msg *FSPMessage) {
	fspgame.LockedFrame.Msgs = append(fspgame.LockedFrame.Msgs, msg)
}

// AddCmdToCurrFrame :
func (fspgame *FSPGame) AddCmdToCurrFrame(cmd int32, cnt string) {
	msg := new(FSPMessage)
	msg.PlayerID = cmd
	msg.Content = cnt
	msg.Cmd = AUTH

	fspgame.LockedFrame.Msgs = append(fspgame.LockedFrame.Msgs, msg)
}

// SetFlag :
func (fspgame *FSPGame) SetFlag(playerID uint32, flag *int16, flagname string) {
	*flag |= (0x01 << (playerID - 1))
	fspgame.logger.Debug("flag name: ", flagname, "value", *flag)
}

// EnterFrame : 驱动游戏状态
func (fspgame *FSPGame) EnterFrame() {
	fspgame.HandleGameState()
}

// HandleGameState :
func (fspgame *FSPGame) HandleGameState() {
	switch fspgame.State {
	case None:
		break
	case Create:
		if fspgame.isFlagFull(fspgame.gameBeginFlag) {
			fspgame.SetGameState(GameBegin, 0, 0)
			fspgame.AddCmdToCurrFrame(GameBegin, "GameBegin")
		}
		break
	case GameBegin:
		// TODO:
		break
	}
}

// OnStateGameCreate : listen gamebegin
func (fspgame *FSPGame) OnStateGameCreate() {
	if fspgame.isFlagFull(fspgame.gameBeginFlag) {
		fspgame.SetGameState(GameBegin, 0, 0)
		fspgame.AddCmdToCurrFrame(GameBegin, "GameBegin")
	}
}

// OnStateGameBegin : listen roundbegin
func (fspgame *FSPGame) OnStateGameBegin() {
	if fspgame.isFlagFull(fspgame.roundBeginFlag) {
		fspgame.SetGameState(RoundBegin, 0, 0)
		// TODO:

		fspgame.AddCmdToCurrFrame(RoundBegin, "RoundBegin")
	}
}

// OnStateRoundBegin : listen controlstart
func (fspgame *FSPGame) OnStateRoundBegin() {
	if fspgame.isFlagFull(fspgame.controlStartFlag) {
		// TODO:
		fspgame.SetGameState(ControlStart, 0, 0)

		fspgame.AddCmdToCurrFrame(ControlStart, "ControlStart")
	}
}

// OnStateControlStart : listen RoundEnd
func (fspgame *FSPGame) OnStateControlStart() {
	// TODO:
}

// SetGameState :
func (fspgame *FSPGame) SetGameState(gamestate, param1, param2 int) {
	fspgame.State = gamestate
	fspgame.stateParam1 = param1
	fspgame.stateParam2 = param2

	fspgame.logger.Debug("game state now :", fspgame.State)
}

// isFlagFull :
func (fspgame *FSPGame) isFlagFull(flag int16) bool {
	if len(fspgame.playerList) > 0 {
		for _, v := range fspgame.playerList {
			if (flag & (0x01 << (v.ID - 1))) == 0 {
				return false
			}
		}
		return true
	}
	return false
}

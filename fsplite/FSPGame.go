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
	roundEndFlag     int16
	gameEndFlag      int16

	onGameExit OnGameExit
	onGameEnd  OnGameEnd

	CurrRoundID int32
	CurrFrameID int32

	LockedFrame           *FSPFrame
	playerList            map[uint32]*FSPPlayer
	playerExitOnNextFrame map[uint32]*FSPPlayer
	logger                *logrus.Entry
}

// OnGameExit : handle game exit
type OnGameExit func(uint32)

// OnGameEnd : handle game end
type OnGameEnd func(int32)

// NewFSPGame : create a fspgame, set state as Create after finish creating
func NewFSPGame(gameid uint32, fspparam *FSPParam) *FSPGame {
	fspgame := new(FSPGame)
	fspgame.gameID = gameid
	fspgame.maxplayerNum = 10
	fspgame.LockedFrame = new(FSPFrame)
	fspgame.CurrRoundID = 0
	fspgame.CurrFrameID = 0
	fspgame.playerList = make(map[uint32]*FSPPlayer)
	fspgame.playerExitOnNextFrame = make(map[uint32]*FSPPlayer)
	fspgame.param = fspparam

	fspgame.logger = logrus.WithFields(logrus.Fields{"GameID": fspgame.gameID})

	fspgame.SetGameState(Create, 0, 0)
	return fspgame
}

// Release : clear data and set state as None
func (fspgame *FSPGame) Release() {
	fspgame.SetGameState(None, 0, 0)
	for _, v := range fspgame.playerList {
		v.Release()
	}
	fspgame.playerList = make(map[uint32]*FSPPlayer)
	fspgame.playerExitOnNextFrame = make(map[uint32]*FSPPlayer)
	fspgame.onGameEnd = nil
	fspgame.onGameExit = nil
}

// AddPlayer : add a player to this game
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
	player.SetAuth(AUTH)
	fspgame.playerList[playerid] = player
	return player
}

// GetPlayer : return a player
func (fspgame *FSPGame) GetPlayer(playerid uint32) *FSPPlayer {
	return fspgame.playerList[playerid]
}

// GetPlayerMap : return map
func (fspgame *FSPGame) GetPlayerMap() map[uint32]*FSPPlayer {
	return fspgame.playerList
}

// OnRecvFromPlayer : listener of player
func (fspgame *FSPGame) OnRecvFromPlayer(player *FSPPlayer, msg *FSPMessage) {
	// handle data player received from session
	fspgame.handleClientCmd(player, msg)
}

// handleClientCmd : handle data
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

		// if we have a OnGameExit Handler, use it
		if fspgame.onGameExit != nil {
			fspgame.onGameExit(player.ID)
		}
	}
}

// AddMsgToCurrFrame : add game related msg to lockedframe
func (fspgame *FSPGame) AddMsgToCurrFrame(playerID uint32, msg *FSPMessage) {
	fspgame.LockedFrame.Msgs = append(fspgame.LockedFrame.Msgs, msg)
}

// AddCmdToCurrFrame : add state related operation to lockedframe
func (fspgame *FSPGame) AddCmdToCurrFrame(cmd int32, cnt string) {
	// TODO: update data format - notify different state info to client(eg. if need send round id to client)
	msg := new(FSPMessage)
	msg.PlayerID = 0
	msg.Content = cnt
	// mean for state
	msg.Cmd = cmd

	fspgame.LockedFrame.Msgs = append(fspgame.LockedFrame.Msgs, msg)
}

// SetFlag : set flag
func (fspgame *FSPGame) SetFlag(playerID uint32, flag *int16, flagname string) {
	*flag |= (0x01 << (playerID - 1))
	fspgame.logger.Debug("flag name: ", flagname, "value", *flag)
}

// EnterFrame : 驱动游戏状态
func (fspgame *FSPGame) EnterFrame() {
	// clear player in the exit queue
	for _, v := range fspgame.playerExitOnNextFrame {
		v.Release()
	}
	// clear playerExitOnNextFrame
	fspgame.playerExitOnNextFrame = make(map[uint32]*FSPPlayer)
	fspgame.HandleGameState()

	if fspgame.State == None {
		return
	}

	if fspgame.LockedFrame.FrameID != 0 || len(fspgame.LockedFrame.Msgs) > 0 {
		for k, v := range fspgame.playerList {
			v.SendToClient(fspgame.LockedFrame)
			if v.WaitForExit {
				fspgame.playerExitOnNextFrame[k] = v
				delete(fspgame.playerList, k)
			}
		}
	}

	// if frameid is 0, means that msg is not game msg. then we redefine a new lockedframe
	if fspgame.LockedFrame.FrameID == 0 {
		fspgame.LockedFrame = new(FSPFrame)
	}

	// only when state is RoundBegin or ControlStart, can we increase CurrFrameID
	if fspgame.State == RoundBegin || fspgame.State == ControlStart {
		fspgame.CurrFrameID++
		fspgame.LockedFrame = new(FSPFrame)
		fspgame.LockedFrame.FrameID = fspgame.CurrFrameID
	}
}

// HandleGameState :
func (fspgame *FSPGame) HandleGameState() {
	switch fspgame.State {
	case None:
		break
	case Create:
		fspgame.OnStateGameCreate()
		break
	case GameBegin:
		fspgame.OnStateGameBegin()
		break
	case RoundBegin:
		fspgame.OnStateRoundBegin()
		break
	case ControlStart:
		fspgame.OnStateControlStart()
		break
	case RoundEnd:
		fspgame.OnStateRoundEnd()
		break
	case GameEnd:
		fspgame.OnStateGameEnd()
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
	// TODO: 是否需要检测玩家的退出情况，如果玩家在游戏回合开始前退出，应该是游戏结束还是继续游戏，只是该玩家无行动而已。

	if fspgame.isFlagFull(fspgame.roundBeginFlag) {
		fspgame.SetGameState(RoundBegin, 0, 0)
		fspgame.IncRoundID()
		fspgame.ClearRound()
		fspgame.AddCmdToCurrFrame(RoundBegin, "RoundBegin")
	}
}

// OnStateRoundBegin : listen controlstart
func (fspgame *FSPGame) OnStateRoundBegin() {
	// TODO: 是否需要检测玩家的退出情况，如果玩家在游戏回合开始前退出，应该是游戏结束还是继续游戏，只是该玩家无行动而已。

	//
	if fspgame.isFlagFull(fspgame.controlStartFlag) {
		fspgame.ResetRoundFlag()
		fspgame.SetGameState(ControlStart, 0, 0)
		fspgame.AddCmdToCurrFrame(ControlStart, "ControlStart")
	}
}

// OnStateControlStart : listen RoundEnd
func (fspgame *FSPGame) OnStateControlStart() {
	// TODO: 是否需要检测玩家的退出情况，如果玩家在游戏回合开始前退出，应该是游戏结束还是继续游戏，只是该玩家无行动而已。

	//
	if fspgame.isFlagFull(fspgame.roundEndFlag) {
		fspgame.SetGameState(RoundEnd, 0, 0)
		fspgame.ClearRound()
		fspgame.AddCmdToCurrFrame(RoundEnd, "RoundEnd")
	}
}

// OnStateRoundEnd :
func (fspgame *FSPGame) OnStateRoundEnd() {
	// TODO: 是否需要检测玩家的退出情况，如果玩家在游戏回合开始前退出，应该是游戏结束还是继续游戏，只是该玩家无行动而已。

	//
	if fspgame.isFlagFull(fspgame.gameEndFlag) {
		// TODO: param1, param2 : 额外信息。
		fspgame.SetGameState(GameEnd, NormalExit, 0)
		fspgame.AddCmdToCurrFrame(GameEnd, "GameEnd")
	}

	if fspgame.isFlagFull(fspgame.roundBeginFlag) {
		fspgame.SetGameState(RoundBegin, 0, 0)
		fspgame.ClearRound()
		fspgame.IncRoundID()
		fspgame.AddCmdToCurrFrame(RoundBegin, "RoundBegin")
	}

}

// OnStateGameEnd :
func (fspgame *FSPGame) OnStateGameEnd() {
	if fspgame.onGameEnd != nil {
		fspgame.onGameEnd(int32(fspgame.stateParam1))
		fspgame.onGameEnd = nil
	}
}

// IsGameEnd : if game is end
func (fspgame *FSPGame) IsGameEnd() bool {
	return fspgame.State == GameEnd
}

// SetGameState : set game state, param1 and param2 is useless now, using 0 is ok.
func (fspgame *FSPGame) SetGameState(gamestate, param1, param2 int) {
	fspgame.State = gamestate
	fspgame.stateParam1 = param1
	fspgame.stateParam2 = param2

	fspgame.logger.Debug("game state now :", fspgame.State)
}

// isFlagFull : check if all player send corresponding msg
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

// CheckGameAbnormalEnd :
// TODO:
func (fspgame *FSPGame) CheckGameAbnormalEnd() {
	return
}

// IncRoundID : increate roundid
func (fspgame *FSPGame) IncRoundID() {
	fspgame.CurrRoundID++
}

// ClearRound : clear lockedframe
func (fspgame *FSPGame) ClearRound() {
	fspgame.LockedFrame = new(FSPFrame)
	fspgame.CurrFrameID = 0
	fspgame.ResetRoundFlag()

	for _, v := range fspgame.playerList {
		if v != nil {
			v.ClearRound()
		}
	}
}

// ResetRoundFlag : reset flag
func (fspgame *FSPGame) ResetRoundFlag() {
	fspgame.roundBeginFlag = 0
	fspgame.controlStartFlag = 0
	fspgame.roundEndFlag = 0
	fspgame.gameEndFlag = 0
}

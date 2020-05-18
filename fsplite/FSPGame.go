package fsplite

import (
	"fmt"
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
	bufferQueue      *Queue

	CurrRoundID int32
	CurrFrameID int32

	UpperController   FSPGameI

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
	fspgame.LockedFrame.Msgs = make([]*FSPMessage,0)
	fspgame.CurrRoundID = 0
	fspgame.CurrFrameID = 0
	fspgame.playerList = make(map[uint32]*FSPPlayer)
	fspgame.playerExitOnNextFrame = make(map[uint32]*FSPPlayer)
	fspgame.param = fspparam
	fspgame.UpperController = nil
	fspgame.bufferQueue = NewQueue()

	fspgame.logger = logrus.WithFields(logrus.Fields{"GameID": fspgame.gameID})

	fspgame.SetGameState(Create, 0, 0)
	return fspgame
}

func (fspgame *FSPGame) GetFlag(flagname string) *int16 {
	switch flagname {
	case "gameBeginFlag":
		return &fspgame.gameBeginFlag
	case "roundBeginFlag":
		return &fspgame.roundBeginFlag
	case "controlStartFlag":
		return &fspgame.controlStartFlag
	case "roundEndFlag":
		return &fspgame.roundEndFlag
	case "gameEndFlag":
		return &fspgame.gameEndFlag
	default:
		return nil
	}
}

// GetGameID:
func (fspgame *FSPGame) GetGameID() uint32 {
	return fspgame.gameID
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

// AddPlayer : add a player to this game, playerid is player's uid, idInGame means the player's id in the single game
func (fspgame *FSPGame) AddPlayer(playerid uint32, session *FSPSession, idInGame uint32) *FSPPlayer {
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

	player := NewFSPPlayer(playerid, idInGame, session, fspgame.param.AuthID, fspgame.OnRecvFromPlayer)
	// player.SetAuth(AUTH)
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
	fmt.Println("Invoke OnRecvFromPlayer")
	// handle data player received from session
	fspgame.handleClientCmd(player, msg)
}

// handleClientCmd : handle data
func (fspgame *FSPGame) handleClientCmd(player *FSPPlayer, msg *FSPMessage) {
	fmt.Println("Invoke handleClientCmd")
	if msg.Cmd == AUTH {
		// fspgame.logger.Debug("auth")
		player.SetAuth(AUTH)
		for fspgame.bufferQueue.Len() >0 {
			fspgame.handleClientMsg(player, fspgame.bufferQueue.Pop().(*FSPMessage))
		}
		return
		// fspgame.logger.Warn("player id", player.ID, "authed: ", player.hasAuthed)
	}
	if !player.HasAuthed() {
		fspgame.logger.Warn("not authed")
		fspgame.bufferQueue.Push(msg)
		return
	}
	fspgame.handleClientMsg(player, msg)
}

func (fspgame *FSPGame) handleClientMsg(player *FSPPlayer, msg *FSPMessage) {
	fspgame.logger.Info("Msg: ",msg)
	switch msg.Cmd {
	case GameBegin:
		fspgame.SetFlag(player.IdInGame, &fspgame.gameBeginFlag, "gamebeginflag")
		fspgame.logger.Info("gamebeginflag value: ", fspgame.gameBeginFlag)
		fspgame.UpperController.OnGameBeginCallBack(player, msg)
	case RoundBegin:
		fspgame.SetFlag(player.IdInGame, &fspgame.roundBeginFlag, "roundbeginflag")
		fspgame.logger.Info("roundbeginflag value: ", fspgame.roundBeginFlag)
		fspgame.UpperController.OnRoundBeginCallBack(player, msg)
	case ControlStart:
		fspgame.SetFlag(player.IdInGame, &fspgame.controlStartFlag, "controlstartflag")
		fspgame.logger.Info("controlstartflag value: ", fspgame.controlStartFlag)
		fspgame.UpperController.OnControlStartCallBack(player, msg)
	case RoundEnd:
		fspgame.SetFlag(player.IdInGame, &fspgame.roundEndFlag, "roundendflag")
		fspgame.logger.Info("roundendflag value: ", fspgame.roundEndFlag)
		//--------Handler RoundEnd msg----------
		fspgame.UpperController.OnRoundEndCallBack(player, msg)
	case GameEnd:
		fspgame.SetFlag(player.IdInGame, &fspgame.gameEndFlag, "gameendflag")
		fspgame.logger.Info("gameendflag value: ", fspgame.gameEndFlag)
	case GameExit:
		fspgame.HandleGameExit(player, msg)
		fspgame.logger.Info("receive gameexit cmd from player id :", player.ID)
	default:
		fspgame.AddMsgToCurrFrame(player.IdInGame, msg)
	}
}

// HandleGameExit :
func (fspgame *FSPGame) HandleGameExit(fspplayer *FSPPlayer, msg *FSPMessage) {
	fspgame.AddMsgToCurrFrame(fspplayer.IdInGame, msg)
	player := fspgame.GetPlayer(fspplayer.ID)
	if player != nil {
		player.WaitForExit = true

		// if we have a OnGameExit Handler, use it
		if fspgame.onGameExit != nil {
			fspgame.onGameExit(player.ID)
		}
	}
}

// AddMsgToCurrFrame : add game related msg to lockedframe
func (fspgame *FSPGame) AddMsgToCurrFrame(idInGame uint32, msg *FSPMessage) {
	fspgame.LockedFrame.Msgs = append(fspgame.LockedFrame.Msgs, msg)
}

// AddCmdToCurrFrame : add state related operation to lockedframe
func (fspgame *FSPGame) AddCmdToCurrFrame(cmd int32, cnt []byte) {
	logrus.Info("AddToCmd: ",cmd)
	msg := new(FSPMessage)
	msg.PlayerID = 0
	msg.Content = cnt
	// mean for state
	msg.Cmd = cmd

	fspgame.LockedFrame.Msgs = append(fspgame.LockedFrame.Msgs, msg)
}

// SetFlag : set flag
func (fspgame *FSPGame) SetFlag(playerIDInGame uint32, flag *int16, flagname string) {
	*flag |= 0x01 << playerIDInGame
	fspgame.logger.Debug("flag name: ", flagname, "value", *flag)
}

// EnterFrame : 驱动游戏状态
func (fspgame *FSPGame) EnterFrame() {
	// clear player in the exit queue
	// fmt.Println("invoke EnterFrame")
	for _, v := range fspgame.playerExitOnNextFrame {
		v.Release()
	}
	// clear playerExitOnNextFrame
	fspgame.playerExitOnNextFrame = make(map[uint32]*FSPPlayer)
	fspgame.HandleGameState()

	if fspgame.State == None {
		return
	}

	if fspgame.LockedFrame.FrameID != 0 || len(fspgame.LockedFrame.Msgs) >= 0 {
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
		fspgame.LockedFrame.Msgs = make([]*FSPMessage,0)
	}

	// only when state is RoundBegin or ControlStart, can we increase CurrFrameID
	if fspgame.State == RoundBegin || fspgame.State == ControlStart {
		fspgame.CurrFrameID++
		fspgame.LockedFrame = new(FSPFrame)
		fspgame.LockedFrame.Msgs = make([]*FSPMessage,0)
		fspgame.LockedFrame.FrameID = fspgame.CurrFrameID
	}
}

// HandleGameState :
func (fspgame *FSPGame) HandleGameState() {
	switch fspgame.State {
	case None:
		break
	case Create:
		fspgame.UpperController.OnStateGameCreate()
		break
	case GameBegin:
		fspgame.UpperController.OnStateGameBegin()
		break
	case RoundBegin:
		fspgame.UpperController.OnStateRoundBegin()
		break
	case ControlStart:
		fspgame.UpperController.OnStateControlStart()
		break
	case RoundEnd:
		fspgame.UpperController.OnStateRoundEnd()
		break
	case GameEnd:
		fspgame.UpperController.OnStateGameEnd()
		break
	}
}

// OnStateGameCreate : listen gamebegin
func (fspgame *FSPGame) OnStateGameCreate() {
	if fspgame.isFlagFull(fspgame.gameBeginFlag) {
		fspgame.SetGameState(GameBegin, 0, 0)
		v := fspgame.UpperController.CreateGameBeginMsg()
		fspgame.AddCmdToCurrFrame(GameBegin, v)
		fspgame.UpperController.OnGameBeginMsgAddCallBack()
	}
}

// OnStateGameBegin : listen roundbegin
func (fspgame *FSPGame) OnStateGameBegin() {
	// TODO: 是否需要检测玩家的退出情况，如果玩家在游戏回合开始前退出，应该是游戏结束还是继续游戏，只是该玩家无行动而已。

	if fspgame.isFlagFull(fspgame.roundBeginFlag) {
		fspgame.SetGameState(RoundBegin, 0, 0)
		fspgame.IncRoundID()
		fspgame.ClearRound()
		//get content via upper FSPGameI interface
		v := fspgame.UpperController.CreateRoundMsg()
		fspgame.AddCmdToCurrFrame(RoundBegin, v)
		fspgame.UpperController.OnRoundBeginMsgAddCallBack()
	}
}

// OnStateRoundBegin : listen controlstart
func (fspgame *FSPGame) OnStateRoundBegin() {
	// TODO: 是否需要检测玩家的退出情况，如果玩家在游戏回合开始前退出，应该是游戏结束还是继续游戏，只是该玩家无行动而已。

	//
	if fspgame.isFlagFull(fspgame.controlStartFlag) {
		fspgame.ResetRoundFlag()
		fspgame.SetGameState(ControlStart, 0, 0)
		// get content via upper FSPGameI interface
		v := fspgame.UpperController.CreateControlStartMsg()
		fspgame.AddCmdToCurrFrame(ControlStart, v)
		fspgame.UpperController.OnControlStartMsgAddCallBack()
	}
}

// OnStateControlStart : listen RoundEnd
func (fspgame *FSPGame) OnStateControlStart() {
	// TODO: 是否需要检测玩家的退出情况，如果玩家在游戏回合开始前退出，应该是游戏结束还是继续游戏，只是该玩家无行动而已。

	//
	if fspgame.isFlagFull(fspgame.roundEndFlag) {
		fspgame.SetGameState(RoundEnd, 0, 0)
		fspgame.ClearRound()
		v := fspgame.UpperController.CreateRoundEndMsg()
		fspgame.AddCmdToCurrFrame(RoundEnd, v)
		fspgame.UpperController.OnRoundEndMsgAddCallBack()
	}
}

// OnStateRoundEnd :
func (fspgame *FSPGame) OnStateRoundEnd() {
	// TODO: 是否需要检测玩家的退出情况，如果玩家在游戏回合开始前退出，应该是游戏结束还是继续游戏，只是该玩家无行动而已。

	//
	if fspgame.isFlagFull(fspgame.gameEndFlag) {
		// TODO: param1, param2 : 额外信息。
		fspgame.SetGameState(GameEnd, NormalExit, 0)
		fspgame.AddCmdToCurrFrame(GameEnd, []byte("GameEnd"))
	}

	if fspgame.isFlagFull(fspgame.roundBeginFlag) {
		fspgame.SetGameState(RoundBegin, 0, 0)
		fspgame.ClearRound()
		fspgame.IncRoundID()
		fspgame.AddCmdToCurrFrame(RoundBegin, []byte("RoundBegin"))
	}

}

// OnStateGameEnd :
func (fspgame *FSPGame) OnStateGameEnd() {
	if fspgame.onGameEnd != nil {
		fspgame.onGameEnd(int32(fspgame.stateParam1))
		fspgame.onGameEnd = nil
	}
}

// OnGameBeginCallBack()
func (fspgame *FSPGame) OnGameBeginCallBack(player *FSPPlayer, message *FSPMessage) {

}

func (fspgame *FSPGame) OnGameBeginMsgAddCallBack() {

}

func (fspgame *FSPGame) CreateGameBeginMsg() []byte {
	return []byte("GameBegin")
}

// OnRoundBeginCallBack()
func (fspgame *FSPGame) OnRoundBeginCallBack(player *FSPPlayer, message *FSPMessage) {

}

func (fspgame *FSPGame) OnRoundBeginMsgAddCallBack() {

}

func (fspgame *FSPGame) CreateRoundMsg() []byte {
	return []byte("RoundBegin")
}

// OnControlStartCallBack()
func (fspgame *FSPGame) OnControlStartCallBack(player *FSPPlayer, message *FSPMessage) {

}

func (fspgame *FSPGame) OnControlStartMsgAddCallBack() {

}

func (fspgame *FSPGame) CreateControlStartMsg() []byte {
	return []byte("ControlStart")
}

// OnRoundEndCallBack()
func (fspgame *FSPGame) OnRoundEndCallBack(player *FSPPlayer, message *FSPMessage)  {

}

func (fspgame *FSPGame) OnRoundEndMsgAddCallBack() {

}

func (fspgame *FSPGame) CreateRoundEndMsg() []byte {
	return []byte("RoundEnd")
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
			if (flag & (0x01 << v.IdInGame)) == 0 {
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

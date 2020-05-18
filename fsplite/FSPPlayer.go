package fsplite

import "github.com/sirupsen/logrus"

// FSPPlayer :
type FSPPlayer struct {
	ID           uint32
	IdInGame     uint32
	session      *FSPSession
	recvListener RecvListener
	hasAuthed    bool
	authid       int32
	frameCache   *Queue
	lastFrameID  int32
	WaitForExit  bool
}

// RecvListener :
type RecvListener func(*FSPPlayer, *FSPMessage)

// NewFSPPlayer : create a new fspplayer, playerID is player's uid
func NewFSPPlayer(playerID uint32,idInGame uint32 , session *FSPSession, authid int32, listener RecvListener) *FSPPlayer {
	fspplayer := new(FSPPlayer)
	fspplayer.ID = playerID
	fspplayer.IdInGame = idInGame
	fspplayer.recvListener = listener
	fspplayer.session = session
	fspplayer.hasAuthed = false
	fspplayer.authid = authid
	fspplayer.frameCache = NewQueue()
	fspplayer.WaitForExit = false
	fspplayer.lastFrameID = 0

	fspplayer.session.SetReceiveListener(fspplayer.OnRecvFromSession)

	return fspplayer
}

// Release :
func (fspplayer *FSPPlayer) Release() {
	if fspplayer.session != nil {
		fspplayer.session.SetReceiveListener(nil)
		fspplayer.session = nil
	}
}

// SendToClient : send fspframe to client
func (fspplayer *FSPPlayer) SendToClient(frame *FSPFrame) {
	if frame != nil {
		// todo - frameCache store fspmesage
		if !fspplayer.frameCache.Contain(frame) || len(frame.Msgs) == 0 {
			fspplayer.frameCache.Push(frame)
		}
	}
	for fspplayer.frameCache.Len() > 0 {
		if fspplayer.sendinterval(fspplayer.frameCache.Peek().(*FSPFrame)) {
			fspplayer.frameCache.Pop()
		}
	}

	// if fspplayer.session != nil {
	// 	fspplayer.session.Send(frame)
	// }
}

// send
func (fspplayer *FSPPlayer) sendinterval(frame *FSPFrame) bool {
	// logrus.Warn("frame.FrameID: ",frame.FrameID, "lastFrameID: ",fspplayer.lastFrameID)
	if frame.FrameID != 0 && frame.FrameID <= fspplayer.lastFrameID {
		if frame.FrameID != 0 {
			logrus.Warn("frame.FrameID: ",frame.FrameID, "lastFrameID: ",fspplayer.lastFrameID)
		}
		return true
	}

	if fspplayer.session != nil {
		if fspplayer.session.Send(frame) {
			fspplayer.lastFrameID = frame.FrameID
			return true
		}
	} else {
		logrus.Warn("session id null in sendinterval")
	}

	return false
}

// OnRecvFromSession : listener of session
func (fspplayer *FSPPlayer) OnRecvFromSession(message *FSPDataC2S) {
	if fspplayer.session.isEndPointChanged {
		fspplayer.hasAuthed = false
		fspplayer.session.isEndPointChanged = false
	}
	if fspplayer.recvListener != nil {
		for _, v := range message.Msgs {
			fspplayer.recvListener(fspplayer, v)
		}
	}
}

// SetAuth :
func (fspplayer *FSPPlayer) SetAuth(auth int32) {
	// todo - 真正的鉴权
	// logrus.Warn("player authid is: ", fspplayer.authid, "auth is: ",auth)
	fspplayer.hasAuthed = auth == fspplayer.authid
	// logrus.Warn("hasauthed: ",fspplayer.hasAuthed)
}

// HasAuthed : check if authed
func (fspplayer *FSPPlayer) HasAuthed() bool {
	return fspplayer.hasAuthed
}

// ISLose :
func (fspplayer *FSPPlayer) ISLose() bool {
	return !fspplayer.session.isActive
}

// ClearRound : clear cache
func (fspplayer *FSPPlayer) ClearRound() {
	fspplayer.frameCache.Clear()
	fspplayer.lastFrameID = 0
}

package ServerManager

import (
	"time"

	"code.holdonbush.top/ServerFramework/IPCWork"
	log "github.com/sirupsen/logrus"
)

type ServerModule struct {
	MInfo  *ServerModuleInfo
	Logger *log.Entry

	// changes
	status    int
	closeChan chan int
	Ipc       *IPCWork.IPCManager
}

func NewServerModule(info *ServerModuleInfo, _logger *log.Entry, _status int, _close chan int, _ipc *IPCWork.IPCManager) *ServerModule {
	servermodule := new(ServerModule)
	servermodule.MInfo = info
	servermodule.Logger = _logger
	servermodule.status = _status
	servermodule.closeChan = _close
	servermodule.Ipc = _ipc
	return servermodule
}

type Server interface {
	GetId() int
	Create()
	GetStatus() int
	Release()
	Start(servern Server)
	Stop()
	Tick()
	GetModuleInfo() *ServerModuleInfo
}

const (
	UnCreated = iota
	Created
	Running
	Stopped
	Released
)

// todo - finish default interface Server's functions

func (server *ServerModule) Tick() {
	server.Logger.Info("Default Tick")
}

func (server *ServerModule) GetId() int {
	return server.MInfo.Id
}

func (server *ServerModule) Create() {
	server.Logger.Info("Default Server Create Called")
	if server.status == UnCreated || server.status == Released {
		server.status = Created
		server.Logger.Info("Server Created")
	}
}

func (server *ServerModule) Start(servern Server) {
	server.Logger.Debug("Start Server ID:", servern.GetId())
	server.Logger.Info("Default Server Started Called")
	if server.status == Running {
		return
	}
	server.status = Running
	go func(servern Server) {
		for true {
			select {
			case _ = <-server.closeChan:
				return
			default:
				// todo - tick func error

				servern.Tick()
				time.Sleep(time.Millisecond)
			}
		}
	}(servern)
	server.Logger.Info("Server Started")
}

func (server *ServerModule) Stop() {
	server.Logger.Info("Default Server Stop Called")
	if server.status == Stopped {
		return
	}
	server.status = Stopped
	server.closeChan <- 1
	server.Logger.Info("Server Stopped")
}

func (server *ServerModule) Release() {
	server.Logger.Info("Default Server Release Called")
	if server.status == Released {
		return
	}
	server.status = Released
	server.Logger.Info("Server Released")
}

func (server *ServerModule) GetModuleInfo() *ServerModuleInfo {
	return server.MInfo
}

func (server *ServerModule) GetStatus() int {
	return server.status
}

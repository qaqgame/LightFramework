package ServerManager

import log "github.com/sirupsen/logrus"

type ServerModule struct {
	MInfo        ServerModuleInfo
	logger       *log.Entry
}

type Server interface {
	GetId() int
	Create()
	GetStatus() int
	Release()
	Start()
	Stop()
	Tick()
	IsCreated() bool
	GetModuleInfo() ServerModuleInfo
}

const (
	UnCreated = iota
	Created
	Running
	Stopped
	Released
)

//
//func (sm *ServerModule) GetId() int {
//	return sm.MInfo.Id
//}
//
//func (sm *ServerModule) Create(info ServerModuleInfo) {
//	sm.MInfo = info
//	sm.logger = log.WithFields(log.Fields{"Server":info.Name})
//	sm.logger.Info("Server Created")
//}
//
//func (sm *ServerModule) Release() {
//	sm.logger.Info("Server Released")
//}
//
//func (sm *ServerModule) Start() {
//	sm.logger.Info("Server Started")
//
//}
//
//func (sm *ServerModule) Stop() {
//	sm.logger.Info("Server Stoped")
//}
//
//func (sm *ServerModule) Tick() {
//
//}
//
//func (sm *ServerModule) GetModuleInfo() ServerModuleInfo{
//	return ServerModuleInfo{}
//}
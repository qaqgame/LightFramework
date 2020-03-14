package ServerManager

import "log"

type ServerModule struct {
	MInfo        ServerModuleInfo
}

func (sm *ServerModule) Id() int {
	return sm.MInfo.Id
}

func (sm *ServerModule) Create(info ServerModuleInfo) {
	sm.MInfo = info
	log.Println(sm.MInfo.Name," Created")
}

func (sm *ServerModule) Release() {
	log.Println(sm.MInfo.Name," Released")
}

func (sm *ServerModule) Start() {
	log.Println(sm.MInfo.Name," Started")
}

func (sm *ServerModule) Stop() {
	log.Println(sm.MInfo.Name," Stopped")
}

func (sm *ServerModule) Tick() {

}
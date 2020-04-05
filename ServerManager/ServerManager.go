package ServerManager

import (
	log "github.com/sirupsen/logrus"
)

type ServerManager struct {
	ServerModules1   map[int]Server
}

func NewServerManager() *ServerManager {
	log.WithFields(log.Fields{
		"Server":"ServerManager",
	}).Info("Create ServerManager")
	smr := new(ServerManager)
	smr.ServerModules1 = make(map[int]Server)
	return smr
}

//func (smr *ServerManager) StartServer(id int) {
//	// log.Println("StartServer ",id)
//	info := GetServerModuleInfo(id)
//	// fullName := Namespace+"."+info.Name+"."+strconv.Itoa(info.Port)
//	logger := log.WithFields(log.Fields{"Server":info.Name})
//	if _,ok := smr.ServerModules[id]; !ok {
//		module := ServerModule{}
//		module.Create(info)
//		smr.ServerModules[id] = &module
//
//		module.Start()
//	} else {
//		smr.ServerModules[id].Start()
//	}
//	logger.Info("Start Server",info.Name)
//}
//
//func (smr *ServerManager) StartAllServer() {
//	v := GetAllServerModuleInfo()
//	for _,s := range v {
//		if _,ok := smr.ServerModules[s.Id]; !ok {
//			module := ServerModule{}
//			module.Create(s)
//			smr.ServerModules[s.Id] = &module
//		} else {
//			smr.ServerModules[s.Id].Start()
//		}
//	}
//	log.WithFields(log.Fields{"Server":"ServerManager"}).Info("Started all servers")
//}
//
//func (smr *ServerManager) StopServer(id int) {
//	if v, ok := smr.ServerModules[id]; ok {
//		v.Stop()
//		v.Release()
//		delete(smr.ServerModules, id)
//	}
//}
//
//func (smr *ServerManager) StopAllServer() {
//	for _,v := range smr.ServerModules {
//		v.Stop()
//		v.Release()
//	}
//	smr.ServerModules = make(map[int]*ServerModule, 0)
//}

func (smr *ServerManager) Tick() {
	// log.Println("Start Tick")
	for _,v := range smr.ServerModules1 {
		v.Tick()
	}
}

func (smr *ServerManager) AddServer(server Server) {
	serverId := server.GetId()
	server.Create()
	smr.ServerModules1[serverId] = server
}

func (smr *ServerManager) RemoveServer(id int) {
	smr.ServerModules1[id].Stop()
	smr.ServerModules1[id].Release()
	delete(smr.ServerModules1,id)
}

func (smr *ServerManager) RemoveAllServer() {
	smr.ServerModules1 = make(map[int]Server)
}

func (smr *ServerManager) StartAServer(id int)  {
	if v, ok := smr.ServerModules1[id]; ok {
		v.Start(&v)
	}
}

func (smr *ServerManager) StartAllServer1() {
	for _,v := range smr.ServerModules1 {
		v.Start(&v)
	}
}

func (smr *ServerManager) StopAServer(id int) {
	server := smr.ServerModules1[id]
	server.Stop()
}

func (smr *ServerManager) StopAllServer1() {
	for _,v := range smr.ServerModules1 {
		v.Stop()
	}
}
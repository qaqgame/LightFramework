package ServerManager

import (
	log "github.com/sirupsen/logrus"
)

type ServerManager struct {
	ServerModules    map[int]*ServerModule
	ServerModules1   map[int]Server
}

func NewServerManager() *ServerManager {
	log.WithFields(log.Fields{
		"Server":"ServerManager",
	}).Info("Create ServerManager")
	smr := new(ServerManager)
	smr.ServerModules = make(map[int]*ServerModule)
	smr.ServerModules1 = make(map[int]Server)
	return smr
}

func (smr *ServerManager) StartServer(id int) {
	// log.Println("StartServer ",id)
	info := GetServerModuleInfo(id)
	// fullName := Namespace+"."+info.Name+"."+strconv.Itoa(info.Port)
	logger := log.WithFields(log.Fields{"Server":info.Name})
	if _,ok := smr.ServerModules[id]; !ok {
		module := ServerModule{}
		module.Create(info)
		smr.ServerModules[id] = &module

		module.Start()
	} else {
		smr.ServerModules[id].Start()
	}
	logger.Info("Start Server",info.Name)
}

func (smr *ServerManager) StartAllServer() {
	v := GetAllServerModuleInfo()
	for _,s := range v {
		if _,ok := smr.ServerModules[s.Id]; !ok {
			module := ServerModule{}
			module.Create(s)
			smr.ServerModules[s.Id] = &module
		} else {
			smr.ServerModules[s.Id].Start()
		}
	}
	log.WithFields(log.Fields{"Server":"ServerManager"}).Info("Started all servers")
}

func (smr *ServerManager) StopServer(id int) {
	if v, ok := smr.ServerModules[id]; ok {
		v.Stop()
		v.Release()
		delete(smr.ServerModules, id)
	}
}

func (smr *ServerManager) StopAllServer() {
	for _,v := range smr.ServerModules {
		v.Stop()
		v.Release()
	}
	smr.ServerModules = make(map[int]*ServerModule, 0)
}

func (smr *ServerManager) Tick() {
	// log.Println("Start Tick")
	for _,v := range smr.ServerModules {
		v.Tick()
	}
}

func (smr *ServerManager) AddServer(server Server) {
	serverId := server.GetId()
	smr.ServerModules1[serverId] = server
}

func (smr *ServerManager) RemoveServer(id int) {
	delete(smr.ServerModules1,id)
}

func (smr *ServerManager) RemoveAllServer() {
	smr.ServerModules1 = make(map[int]Server)
}

func (smr *ServerManager) StartAServer(id int)  {
	if v, ok := smr.ServerModules1[id]; ok {
		if v.IsCreated() {
			v.Start()
		} else {
			v.Create(smr.ServerModules1[id].GetModuleInfo())
			v.Start()
		}
	}
}

func (smr *ServerManager) StartAllServer1() {
	for k,v := range smr.ServerModules1 {
		if v.IsCreated() {
			v.Start()
		} else {
			v.Create(smr.ServerModules1[k].GetModuleInfo())
			v.Start()
		}
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
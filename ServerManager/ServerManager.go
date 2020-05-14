package ServerManager

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

type ServerManager struct {
	ServerModules1 map[int]Server
}

func NewServerManager() *ServerManager {
	log.WithFields(log.Fields{
		"Server": "ServerManager",
	}).Info("Create ServerManager")
	smr := new(ServerManager)
	smr.ServerModules1 = make(map[int]Server)
	return smr
}


func (smr *ServerManager) Tick() {
	// log.Println("Start Tick")
	for _, v := range smr.ServerModules1 {
		v.Tick()
	}
}

func (smr *ServerManager) AddServer(server Server) {
	serverId := server.GetId()
	fmt.Println("ServerID: ",serverId)
	server.Create()
	smr.ServerModules1[serverId] = server
}

func (smr *ServerManager) RemoveServer(id int) {
	smr.ServerModules1[id].Stop()
	smr.ServerModules1[id].Release()
	delete(smr.ServerModules1, id)
}

func (smr *ServerManager) RemoveAllServer() {
	smr.ServerModules1 = make(map[int]Server)
}

func (smr *ServerManager) StartAServer(id int) {
	if v, ok := smr.ServerModules1[id]; ok {
		v.Start(v)
	}
}

func (smr *ServerManager) StartAllServer1() {
	for _, v := range smr.ServerModules1 {
		v.Start(v)
	}
}

func (smr *ServerManager) StopAServer(id int) {
	server := smr.ServerModules1[id]
	server.Stop()
}

func (smr *ServerManager) StopAllServer1() {
	for _, v := range smr.ServerModules1 {
		v.Stop()
	}
}

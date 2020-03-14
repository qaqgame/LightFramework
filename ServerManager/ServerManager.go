package ServerManager

import "log"

type ServerManager struct {
	ServerModules    map[int]ServerModule
}

func (smr *ServerManager) Init() {
	log.Println("ServerManager Init")
	smr.ServerModules = make(map[int]ServerModule)
}

func (smr *ServerManager) StartServer(id int) {
	log.Println("StartServer ",id)
	info := GetServerModuleInfo(id)
	// fullName := Namespace+"."+info.Name+"."+strconv.Itoa(info.Port)

	if _,ok := smr.ServerModules[id]; !ok {
		module := ServerModule{}
		module.Create(info)
		smr.ServerModules[id] = module

		module.Start()
	}
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
	smr.ServerModules = make(map[int]ServerModule, 0)
}

func (smr *ServerManager) Tick() {
	log.Println("Start Tick")
	for _,v := range smr.ServerModules {
		v.Tick()
	}
}
package ServerManager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type ServerModuleInfo struct {
	Id   int
	Name string
	Port int
}

type ConfigUnMarshal struct {
	ServerModuleInfos  []ServerModuleInfo
}

const (
	filename = "ServerConfig.json"
)


var (
	Namespace = "LFW.Server"
	mapServerModuleInfo = make(map[int]ServerModuleInfo)
)

func GetServerModuleInfo(id int) ServerModuleInfo {
	log.Println("GetServerModuleInfo")
	if len(mapServerModuleInfo) == 0 {
		readConfig()
	}

	return mapServerModuleInfo[id]
}

func readConfig() {
	log.Println("Read Config file")
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		fmt.Println("readConfig error:",err)
	}

	cnt, err := ioutil.ReadAll(f)
	serverinfos := ConfigUnMarshal{}
	err = json.Unmarshal(cnt, &serverinfos)
	if err != nil {
		fmt.Println("readConfig error:",err)
	}

	for _,v := range serverinfos.ServerModuleInfos {
		mapServerModuleInfo[v.Id] = v
	}
}
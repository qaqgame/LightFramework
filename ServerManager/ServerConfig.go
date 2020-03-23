package ServerManager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type ServerModuleInfo struct {
	Id   int     `json:"Id"`
	Name string  `json:"Name"`
	Port int     `json:"Port"`
}

type ConfigUnMarshal struct {
	ServerModuleInfos  []ServerModuleInfo     `json:"ConfigUnMarshal"`
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

func GetAllServerModuleInfo() []ServerModuleInfo {
	if len(mapServerModuleInfo) == 0 {
		readConfig()
	}
	ans := make([]ServerModuleInfo,len(mapServerModuleInfo))
	for k,v := range mapServerModuleInfo{
		ans[k] = v
	}
	return ans
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
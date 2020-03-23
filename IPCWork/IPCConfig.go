package IPCWork

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type IPCInfo struct {
	Id   int `json:"Id"`
	Port int `json:"Port"`
}

type IPCInfoUnMarshal struct {
	IPCInfos      []IPCInfo  `json:"ServerModuleInfos"`
}

const (
	filename = "ServerConfig.json"
)

var (
	mapIPCInfo = make(map[int]IPCInfo)
)

func GetIPCInfo(id int) IPCInfo {
	if len(mapIPCInfo) == 0 {
		readConfig()
	}
	return mapIPCInfo[id]
}

func readConfig() {
	log.Println("Read Config file")
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		fmt.Println("readConfig error:",err)
	}

	cnt, err := ioutil.ReadAll(f)
	fmt.Println("read: ",string(cnt))
	IPC := IPCInfoUnMarshal{}
	err = json.Unmarshal(cnt, &IPC)
	if err != nil {
		fmt.Println("readConfig error:",err)
	}
	fmt.Println("len",len(IPC.IPCInfos))
	for _,v := range IPC.IPCInfos {
		mapIPCInfo[v.Id] = v
		fmt.Println(v.Id, v.Port)
	}
}


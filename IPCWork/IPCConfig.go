package IPCWork

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type IPCInfo struct {
	Id   int
	Port int
}

type IPCInfoUnMarshal struct {
	IPCInfos      []IPCInfo
}

const (
	filename = "IPCConfig.json"
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
	IPCInfo := IPCInfoUnMarshal{}
	err = json.Unmarshal(cnt, &IPCInfo)
	if err != nil {
		fmt.Println("readConfig error:",err)
	}

	for _,v := range IPCInfo.IPCInfos {
		mapIPCInfo[v.Id] = v
	}
}


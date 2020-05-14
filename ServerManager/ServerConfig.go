package ServerManager

import (
	"code.holdonbush.top/ServerFramework/common"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	filename = "ServerConfig.json"
)


var (
	mapServerModuleInfo = make(map[int]common.ServerModuleInfo)
)

func GetServerModuleInfo(id int) common.ServerModuleInfo {
	log.Println("GetServerModuleInfo")
	if len(mapServerModuleInfo) == 0 {
		readConfig()
	}

	return mapServerModuleInfo[id]
}

func GetAllServerModuleInfo() []common.ServerModuleInfo {
	if len(mapServerModuleInfo) == 0 {
		readConfig()
	}

	ans := make([]common.ServerModuleInfo,len(mapServerModuleInfo))
	for k,v := range mapServerModuleInfo{
		ans[k-1] = v
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
	serverinfos := common.ConfigUnMarshal{}
	err = json.Unmarshal(cnt, &serverinfos)
	if err != nil {
		fmt.Println("readConfig error:",err)
	}

	for _,v := range serverinfos.ServerModuleInfos {
		mapServerModuleInfo[v.Id] = v
	}
}
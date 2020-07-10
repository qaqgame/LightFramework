package ServerFramework

import (
	"code.holdonbush.top/ServerFramework/ServerManager"
	"log"
	"time"
)

func Run(smr *ServerManager.ServerManager) {
	log.Println("Start MainLoop, 服务端已启动.....")

	for {
		time.Sleep(time.Millisecond)
	}
}

package ServerFramework

import (
	"code.holdonbush.top/ServerFramework/ServerManager"
	"log"
	"time"
)

func Run(smr *ServerManager.ServerManager) {
	log.Println("Start MainLoop")

	for (true) {
		smr.Tick()
		time.Sleep(1*time.Second)
	}
}

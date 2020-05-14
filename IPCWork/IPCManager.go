package IPCWork

import (
	"code.holdonbush.top/ServerFramework/common"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type IPCManager struct {
	myId       int
	myPort     int
	isRunning  bool
	logger     *log.Entry
	stopSignal chan int
}

func NewIPCManager(minfo *common.ServerModuleInfo) *IPCManager {
	ipc := new(IPCManager)
	ipc.myId = minfo.Id
	ipc.myPort = minfo.Port
	fmt.Println("MYPORT", ipc.myPort)
	ipc.logger = log.WithFields(log.Fields{"Server": "IPCManager of" + strconv.Itoa(ipc.myId)})
	ipc.isRunning = false
	ipc.stopSignal = make(chan int, 2)

	return ipc
}

func (ipc *IPCManager) RegisterRPC(m interface{}) {
	rpc.Register(m)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp4", ":"+strconv.Itoa(ipc.myPort))
	ipc.logger.Info("port is", ipc.myPort)
	if e != nil {
		ipc.logger.Warn("Listen tcp error", e)
	}
	go http.Serve(l, nil)
}

func (ipc *IPCManager) Clean() {
	ipc.Stop()
}

func (ipc *IPCManager) Start() {

}

func (ipc *IPCManager) Stop() {
	ipc.stopSignal <- 1
}

func (ipc *IPCManager) CallRpc(args, reply interface{}, port int, rpcname string) bool {
	c, err := rpc.DialHTTP("tcp4", "127.0.0.1:"+strconv.Itoa(port))
	defer c.Close()
	if err != nil {
		ipc.logger.Warn("Dial tcp error:", err)
	}

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}
	if err.Error() != "unexpected EOF" {
		ipc.logger.Warn("rpc call error:", err)
	}
	return false
}

func (ipc *IPCManager) CallRpcAsync(args, reply interface{}, port int, rpcname string) *rpc.Call{
	c, err := rpc.DialHTTP("tcp4", "127.0.0.1:"+strconv.Itoa(port))
	defer c.Close()
	if err != nil {
		ipc.logger.Warn("Dial tcp error:", err)
	}

	v := c.Go(rpcname, args, reply, nil)
	return v
}

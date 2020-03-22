package IPCWork

import "reflect"

// 使用GOB包编解码即可
type IPCMessage struct {
	Src    int
	RPCMsg RPCMsgInIPC
}

type RPCMsgInIPC struct {
	Name    string
	Args    []RPCMsgInIPCArgs
}

type RPCMsgInIPCArgs struct {
	ArgType      reflect.Type
	ArgValue     reflect.Value
}
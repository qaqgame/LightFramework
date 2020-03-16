package Server

import (
	"code.holdonbush.top/ServerFramework/Network"
	"fmt"
	"reflect"
	"testing"
)

func TestNetManager_HandleRPCMessage(t *testing.T) {
	kcp := NewKCPSession(1,nil,nil)
	n := NewNetManager(222)
	rpc := Network.RPCMessage{Name:"222"}
	n.HandleRPCMessage(kcp,&rpc)
	fmt.Println(reflect.TypeOf(kcp).String())
}
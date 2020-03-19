package Server

import (
	"code.holdonbush.top/ServerFramework/Network"
	"fmt"
	"github.com/golang/protobuf/proto"
	"testing"
)

func TestNetManager_HandleRPCMessage(t *testing.T) {

	t1 := new(Network.ProtocolHead)
	t1.Cmd = 1
	t1.UId = 2
	t1.CheckSum = 3
	t1.DataSize = 4
	t1.Index = 5

	v,_ := proto.Marshal(t1)
	t2 := new(Network.ProtocolHead)
	proto.Unmarshal(v, t2)
	fmt.Println(t2)

}
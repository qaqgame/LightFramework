package Network

import (
	"github.com/golang/protobuf/proto"
	"log"
)

func SerializeNetMsg(nm *NetMessage) []byte {
	out,err := proto.Marshal(nm)
	if err != nil {
		log.Fatal("SerializeNetMsg err: ",err)
	}
	return out
}

func SerializeRPCMsg(rm *RPCMessage) []byte {
	out,err := proto.Marshal(rm)
	if err != nil {
		log.Fatal("SerializeNetMsg err: ",err)
	}
	return out
}

func DeserializeNetMsg(buf []byte) *NetMessage {
	nm := &NetMessage{}
	if err := proto.Unmarshal(buf, nm); err != nil {
		log.Fatal("DeserializeNetMsg err: ",err)
	}
	return nm
}

func DeserializeRPCMsg(buf []byte) *RPCMessage {
	rm := &RPCMessage{}
	if err := proto.Unmarshal(buf, rm); err != nil {
		log.Fatal("DeserializeNetMsg err: ",err)
	}
	return rm
}
package Network

import (
	"encoding/binary"
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
	if buf == nil {
		log.Println("nil")
	}
	//ds := int32(binary.LittleEndian.Uint32(buf[12:16]))
	//d1 := make([]byte,4)
	//binary.BigEndian.PutUint32(d1,uint32(ds))
	//tmp := make([]byte,len(buf))
	//tmp = append(buf[:12],d1...)
	//tmp = append(tmp,buf[16:]...)
	log.Println(binary.BigEndian.Uint32(buf[:4]),int32(binary.BigEndian.Uint32(buf[12:16])))
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
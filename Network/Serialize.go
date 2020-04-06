package Network

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
)

func SerializeNetMsg(nm *NetMessage) []byte {
	//out,err := proto.Marshal(nm)
	//if err != nil {
	//	log.Fatal("SerializeNetMsg err: ",err)
	//}
	buf := make([]byte,0)
	buf2 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf2,nm.Head.UId)
	buf= append(buf, buf2...)
	binary.BigEndian.PutUint32(buf2,nm.Head.Cmd)
	buf= append(buf, buf2...)
	binary.BigEndian.PutUint32(buf2,nm.Head.Index)
	buf= append(buf, buf2...)
	binary.BigEndian.PutUint32(buf2,nm.Head.DataSize)
	buf= append(buf, buf2...)
	binary.BigEndian.PutUint32(buf2,nm.Head.CheckSum)
	buf= append(buf, buf2...)
	buf = append(buf, nm.Content...)

	log.Println(time.Now().Unix(),buf)

	return buf
}

func SerializeRPCMsg(rm proto.Message) []byte {
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
	nm.Head = &ProtocolHead{}
	//if err := proto.Unmarshal(buf, nm); err != nil {
	//	log.Fatal("DeserializeNetMsg err: ",err)
	//}
	nm.Head.UId = binary.BigEndian.Uint32(buf[:4])
	nm.Head.Cmd = binary.BigEndian.Uint32(buf[4:8])
	nm.Head.Index = binary.BigEndian.Uint32(buf[8:12])
	// todo
	nm.Head.DataSize = binary.BigEndian.Uint32(buf[12:16])
	nm.Head.CheckSum = binary.BigEndian.Uint32(buf[16:20])
	nm.Content = buf[20:]
	return nm
}

func DeserializeRPCMsg(buf []byte) *RPCMessage {
	rm := &RPCMessage{}
	if err := proto.Unmarshal(buf, rm); err != nil {
		log.Fatal("DeserializeNetMsg err: ",err)
	}
	return rm
}
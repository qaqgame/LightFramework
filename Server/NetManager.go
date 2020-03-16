package Server

import (
	"code.holdonbush.top/ServerFramework/Network"
	"github.com/golang/protobuf/proto"
	"log"
	"reflect"
)

type NetManager struct {
	gateway           *Gateway
	rpc               *Network.RPCManager
	authCmd           uint32
	lastRPCMethod     string
	lastRPCISession   ISession
	mapListenerHelper map[uint32]*ListenerHelper
}

func NewNetManager(port int) *NetManager {
	n := new(NetManager)
	g := NewGateway(port, n)
	n.gateway = g
	n.rpc = new(Network.RPCManager)
	n.mapListenerHelper = make(map[uint32]*ListenerHelper)

	return n
}

func (netManager *NetManager) Clean() {
	if netManager.gateway != nil {
		netManager.gateway.Clean()
		netManager.gateway = nil
	}

	if netManager.rpc != nil {
		netManager.rpc.Clean()
		netManager.rpc = nil
	}
}

func (netManager *NetManager) Tick() {
	netManager.Tick()
}

func (netManager *NetManager) OnReceive(session ISession, bytes []byte, length int) {
	msg := Network.DeserializeNetMsg(bytes)

	if session.IsAuth() {
		if msg.Head.Cmd == 0 {
			rpcmsg := Network.DeserializeRPCMsg(msg.Content)
			netManager.HandleRPCMessage(session,rpcmsg)
		} else {
			netManager.HandlePBMessage(session,msg)
		}
	} else {
		if msg.Head.Cmd == netManager.authCmd {
			netManager.HandlePBMessage(session,msg)
		} else {
			log.Println("UnAuth cmd message")
		}
	}
}

func (netManager *NetManager) SetAuthCmd(cmd uint32) {
	netManager.authCmd = cmd
}


// RPC
func (netManager *NetManager) HandleRPCMessage(session ISession, rpc *Network.RPCMessage) {
	m := netManager.rpc.GetMethod(session, rpc.Name)
	if m != nil {
		in := make([]reflect.Value, len(rpc.RPCRawArgs)+1)
		in[0] = reflect.ValueOf(session)

		v := rpc.GetArgs()
		rawArgs := rpc.RPCRawArgs
		if len(v) == (m.Type.NumIn()-1) {
			for i := 0; i < len(rawArgs); i++ {
				if rawArgs[i].RawValueType == Network.RPCArgType_PBObject {
					t := m.Type.In(i+1)
					err := proto.Unmarshal(rawArgs[i].RawValue, t.(proto.Message))
					if err != nil {
						log.Println("error handlerrpcmessage: ",err)
					}
					v[i] = reflect.ValueOf(t.(proto.Message))
				}
				in[i+1] = v[i]
			}
			// in = append(in,v...)
			netManager.lastRPCMethod = m.Name
			netManager.lastRPCISession = session
			m.Func.Call(in)
			netManager.lastRPCISession = nil
			netManager.lastRPCMethod = ""
		} else {
			log.Println("parameters num is not same")
		}
	} else {
		log.Println("message is not exist")
	}
}



// 服务端向客户端发送
func (netManager *NetManager) Invoke(session ISession, name string, args ...interface{}) {
	rpcmsg := Network.RPCMessage{}
	rpcmsg.Name = name
	rpcmsg.SetArgs(args)

	buf := Network.SerializeRPCMsg(&rpcmsg)

	netmsg := Network.NetMessage{}
	netmsg.Head = &Network.ProtocolHead{}
	netmsg.Head.DataSize = uint32(len(buf))
	netmsg.Content = buf

	sendv := Network.SerializeNetMsg(&netmsg)
	session.Send(sendv,len(sendv))
}

func (netManager *NetManager) InvokeBroadCast(sessions []ISession, name string, args ...interface{}) {
	rpcmsg := Network.RPCMessage{}
	rpcmsg.Name = name
	rpcmsg.SetArgs(args)

	buf := Network.SerializeRPCMsg(&rpcmsg)

	netmsg := Network.NetMessage{}
	netmsg.Head = &Network.ProtocolHead{}
	netmsg.Head.DataSize = uint32(len(buf))
	netmsg.Content = buf

	sendv := Network.SerializeNetMsg(&netmsg)
	for _,v := range sessions {
		v.Send(sendv,len(sendv))
	}
}

func (netManager *NetManager) Return(args ...interface{}) {
	name := "On" + netManager.lastRPCMethod

	rpcmsg := Network.RPCMessage{}
	rpcmsg.Name = name
	rpcmsg.SetArgs(args)

	buf := Network.SerializeRPCMsg(&rpcmsg)

	netmsg := Network.NetMessage{}
	netmsg.Head = &Network.ProtocolHead{}
	netmsg.Head.DataSize = uint32(len(buf))
	netmsg.Content = buf

	sendv := Network.SerializeNetMsg(&netmsg)
	netManager.lastRPCISession.Send(sendv,len(sendv))
}

func (netManager *NetManager) RegisterRPCListener(listener interface{}) {
	netManager.rpc.RegisterObj(listener)
}

func (netManager *NetManager) UnRegisterRPCListener(listener interface{}) {
	netManager.rpc.UnRegisterObj(listener)
}

type ListenerHelper struct {
	TMsg        proto.Message
	onMsg       OnMsg
}

type OnMsg func(session ISession, index uint32, tmsg proto.Message)

// Proto
func (netManager *NetManager) HandlePBMessage(session ISession, pb *Network.NetMessage) {
	helper := netManager.mapListenerHelper[pb.Head.Cmd]
	if helper != nil {
		obj := helper.TMsg
		_ = proto.Unmarshal(pb.Content, obj)
		if obj != nil {
			//in := []reflect.Value{reflect.ValueOf(session),reflect.ValueOf(pb.Head.Index),reflect.ValueOf(obj)}
			//helper.onMsg.Func.Call(in)
			helper.onMsg(session, pb.Head.Index, obj)
		}
	} else {
		log.Println("no listener")
	}
}

func (netManager *NetManager) Send(session ISession, index, cmd uint32, msg proto.Message) {
	netmsg := Network.NetMessage{}
	netmsg.Head.Index = index
	netmsg.Head.Cmd = cmd
	netmsg.Head.UId = session.GetUid()
	netmsg.Content,_ = proto.Marshal(msg)
	netmsg.Head.DataSize = uint32(len(netmsg.Content))

	buf := Network.SerializeNetMsg(&netmsg)

	session.Send(buf,len(buf))
}

func (netManager *NetManager) AddListener(cmd uint32, onmsg OnMsg, tmsg proto.Message) {

	helper := ListenerHelper{
		TMsg:  tmsg,
		onMsg: onmsg,
	}
	netManager.mapListenerHelper[cmd] = &helper
}
package Server

import (
	"code.holdonbush.top/ServerFramework/Network"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"reflect"
)

type NetManager struct {
	Gateway           *Gateway
	rpc               *Network.RPCManager
	authCmd           uint32
	lastRPCMethod     string
	lastRPCISession   ISession
	mapListenerHelper map[uint32]*ListenerHelper
	logger            *log.Entry
}

func NewNetManager(port int,arg ...interface{}) *NetManager {
	//log.Println("new NetManager")

	n := new(NetManager)
	if len(arg) == 1 {
		n.logger = arg[0].(*log.Entry)
	}
	g := NewGateway(port, n)
	n.Gateway = g
	n.rpc = new(Network.RPCManager)
	n.mapListenerHelper = make(map[uint32]*ListenerHelper)
	n.logger.Info("NetManager Created")
	return n
}

func (netManager *NetManager) GetLogger() *log.Entry {
	return netManager.logger
}

func (netManager *NetManager) Clean() {
	if netManager.Gateway != nil {
		netManager.Gateway.Clean()
		netManager.Gateway = nil
	}

	if netManager.rpc != nil {
		netManager.rpc.Clean()
		netManager.rpc = nil
	}
}

func (netManager *NetManager) Tick() {
	netManager.Gateway.Tick()
}

func (netManager *NetManager) OnReceive(session ISession, bytes []byte, length int) {
	//log.Println("onreceive buf lenght: ",len(bytes))
	netManager.logger.Debug("OnReceive in NetManager, receive data len: ",len(bytes))
	msg := Network.DeserializeNetMsg(bytes)

	if session.IsAuth() {
		if msg.Head.Cmd == 0 {
			rpcmsg := Network.DeserializeRPCMsg(msg.Content)
			netManager.HandleRPCMessage(session,rpcmsg)
		} else {
			//log.Println("Use Proto Handler")
			netManager.logger.Debug("OnReceive in NetManager, Use HandlePBMessage")
			netManager.HandlePBMessage(session,msg)
		}
	} else {
		if msg.Head.Cmd == netManager.authCmd {
			netManager.HandlePBMessage(session,msg)
		} else {
			//log.Println("UnAuth cmd message")
			netManager.logger.Debug("OnReceive in NetManager, UnAuth cmd message")
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
						//log.Println("error handlerrpcmessage: ",err)
						netManager.logger.Error("HandleRPCMessage in NetManager, error Unmarshal: ",err)
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
			//log.Println("parameters num is not same")
			netManager.logger.Debug("HandleRPCMessage in NetManager, RPC function's parameters num is not same")
		}
	} else {
		//log.Println("message is not exist")
		netManager.logger.Debug("HandleRPCMessage in NetManage, Method not found")
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

func (netManager *NetManager) RegisterRPCMethod(listener interface{}, name string) {
	netManager.rpc.RegisterMethod(listener,name)
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
		//fmt.Println("HandlePBMessge: ",helper.TMsg,obj,obj.(proto.Message),reflect.TypeOf(obj),reflect.TypeOf(obj.(proto.Message)))
		netManager.logger.Debug("HandlePBMessage in NetManager, TMsg type is",obj)
		err := proto.Unmarshal(pb.Content, obj)
		if err != nil {
			//log.Println("unmarshal content error: ",err)
			netManager.logger.Warn("HandlePBMessage in NetManager, Unmarshal content error:",err)
		}
		//fmt.Println("unmarshaled: ",obj)
		netManager.logger.Debug("HandlePBMessage in NetManager, Unmarshal result:", obj)
		if obj != nil {
			//in := []reflect.Value{reflect.ValueOf(session),reflect.ValueOf(pb.Head.Index),reflect.ValueOf(obj)}
			//helper.onMsg.Func.Call(in)
			helper.onMsg(session, pb.Head.Index, obj)
		}
	} else {
		//log.Println("no listener")
		netManager.logger.Debug("HandlePBMessage in NetManager, listener not found")
	}
}

func (netManager *NetManager) Send(session ISession, index, cmd uint32, msg proto.Message) {
	netmsg := Network.NetMessage{}
	netmsg.Head = &Network.ProtocolHead{}
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
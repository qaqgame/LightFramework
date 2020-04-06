package Network

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRPCManager_GetMethod(t *testing.T) {
	rmr := NewRPCManager()
	rmr.RegisterMethod(rmr,"TestMethod", reflect.ValueOf(rmr))
	//rmr.RegisterObj(rmr) 
	//fmt.Println(rmr.mListListener)
	m := rmr.GetMethod("*Network.RPCManager", "TestMethod")
	in := []reflect.Value{reflect.ValueOf(rmr),reflect.ValueOf(1),reflect.ValueOf(" haha"),reflect.ValueOf(true)}
	//fmt.Println("v",m.Type.NumIn(),m.Type.In(2),m.Name)
	//
	v := m.Func.Call(in)
	fmt.Println(v[0])
}

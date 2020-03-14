package Network

import (
	"reflect"
	"testing"
)

func TestRPCManager_GetMethod(t *testing.T) {
	rmr := RPCManager{}
	rmr.Init()
	rmr.Register(rmr,"TestMethod")
	m := rmr.GetMethod(rmr, "TestMethod")
	in := []reflect.Value{reflect.ValueOf(rmr),reflect.ValueOf(1)}
	m.Func.Call(in)
}

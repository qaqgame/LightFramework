package Network

import (
	"fmt"
	"reflect"
)

type RPCManager struct {
	mListListener     map[string][]*reflect.Method
	receiverRegister  map[string]reflect.Value
}

func NewRPCManager() *RPCManager {
	rmr := new(RPCManager)
	rmr.mListListener = make(map[string][]*reflect.Method)
	rmr.receiverRegister = make(map[string]reflect.Value)
	return rmr
}

func (rmr *RPCManager) Clean() {
	rmr.receiverRegister = make(map[string]reflect.Value)
	rmr.mListListener = make(map[string][]*reflect.Method)
}

func (rmr *RPCManager) RegisterMethod(listener interface{}, methodName string, structtype reflect.Value) {
	t := reflect.TypeOf(listener)
	m, _ := t.MethodByName(methodName)
	
	fmt.Println("register name",t.String())

	rmr.receiverRegister[t.String()] = structtype

	if m.Name == "" {
		return
	}
	if rmr.mListListener[t.String()] == nil {
		rmr.mListListener[t.String()] = make([]*reflect.Method, 0)
	}
	rmr.mListListener[t.String()] = append(rmr.mListListener[t.Name()], &m)
}

func (rmr *RPCManager) RegisterMethods(listener interface{}, structtype reflect.Value, methodNames ...string) {
	t := reflect.TypeOf(listener)

	rmr.receiverRegister[t.String()] = structtype

	for i:=0; i<len(methodNames); i++ {
		m,_ := t.MethodByName(methodNames[i])
		if m.Name == "" {
			continue
		}
		if rmr.mListListener[t.String()] == nil {
			rmr.mListListener[t.String()] = make([]*reflect.Method, 0)
		}
		rmr.mListListener[t.String()] = append(rmr.mListListener[t.String()], &m)
	}
}

func (rmr *RPCManager) RegisterObj(listener interface{}, structtype reflect.Value) {
	t := reflect.TypeOf(listener)
	rmr.receiverRegister[t.String()] = structtype
	rmr.mListListener[t.String()] = make([]*reflect.Method, 0)
}

func (rmr *RPCManager) UnRegisterObj(listener interface{}) {
	delete(rmr.receiverRegister, reflect.TypeOf(listener).String())
	delete(rmr.mListListener, reflect.TypeOf(listener).String())
}

func (rmr *RPCManager) UnRegisterMethod(listener interface{}, name string) {
	index := 0
	objType := reflect.TypeOf(listener).String()
	for k, v := range rmr.mListListener[objType] {
		if v.Name == name {
			index = k
			break
		}
	}
	rmr.mListListener[objType] = append(rmr.mListListener[objType][0:index], rmr.mListListener[objType][index+1:]...)
}

func (rmr *RPCManager) GetMethod(listener string, methodName string) *reflect.Method {
	for _, v := range rmr.mListListener[listener] {
		fmt.Println(v.Name)
		if v.Name == methodName {
			return v
		}
	}
	return nil
}

// GetRegisterValue
func (rmr *RPCManager) GetRegisterValue(key string) reflect.Value {
	return rmr.receiverRegister[key]
}







func (rmr *RPCManager) TestMethod(v int, s string, b bool) int {
	fmt.Println("here is in TestMethod ", v, s)
	return 1
}


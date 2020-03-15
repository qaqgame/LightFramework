package Network

import (
	"fmt"
	"reflect"
)


type RPCManager struct {
	mListListener      map[string][]reflect.Method
}

func NewRPCManager() *RPCManager {
	rmr := new(RPCManager)
	rmr.mListListener = make(map[string][]reflect.Method,0)
	return rmr
}

func (rmr *RPCManager) Clean() {
	rmr.mListListener = make(map[string][]reflect.Method,0)
}

func (rmr *RPCManager) Register(listener interface{}, methodName string) {
	t := reflect.TypeOf(listener)
	m, _ := t.MethodByName(methodName)
	if m.Name == "" {
		return
	}
	rmr.mListListener[t.String()] = append(rmr.mListListener[t.Name()], m)
}

func (rmr *RPCManager) UnRegister(listenner interface{}) {
	delete(rmr.mListListener,reflect.TypeOf(listenner).String())
}

func (rmr *RPCManager) GetMethod(listener interface{},methodName string) reflect.Method {
	for _,v := range rmr.mListListener[reflect.TypeOf(listener).String()] {
		fmt.Println(v.Name)
		if v.Name == methodName {
			return v
		}
	}
	return reflect.Method{}
}

func (rmr RPCManager) TestMethod(v int) {
	fmt.Println("here is in TestMethod ",v)
}
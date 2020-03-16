package Network

import (
	"fmt"
	"reflect"
)


type RPCManager struct {
	mListListener      map[string][]*reflect.Method
}

func NewRPCManager() *RPCManager {
	rmr := new(RPCManager)
	rmr.mListListener = make(map[string][]*reflect.Method,0)
	return rmr
}

func (rmr *RPCManager) Clean() {
	rmr.mListListener = make(map[string][]*reflect.Method,0)
}

func (rmr *RPCManager) RegisterMethod(listener interface{}, methodName string) {
	t := reflect.TypeOf(listener)
	m, _ := t.MethodByName(methodName)
	if m.Name == "" {
		return
	}
	rmr.mListListener[t.String()] = append(rmr.mListListener[t.Name()], &m)
}

func (rmr *RPCManager) RegisterObj(listener interface{}) {
	t := reflect.TypeOf(listener)
	rmr.mListListener[t.String()] = nil
}

func (rmr *RPCManager) UnRegisterObj(listener interface{}) {
	delete(rmr.mListListener,reflect.TypeOf(listener).String())
}

func (rmr *RPCManager) UnRegisterMethod(listener interface{}, name string) {
	index := 0
	objType := reflect.TypeOf(listener).String()
	for k,v := range rmr.mListListener[objType] {
		if v.Name == name {
			index = k
			break
		}
	}
	rmr.mListListener[objType] = append(rmr.mListListener[objType][0:index], rmr.mListListener[objType][index+1:]...)
}

func (rmr *RPCManager) GetMethod(listener interface{},methodName string) *reflect.Method {
	for _,v := range rmr.mListListener[reflect.TypeOf(listener).String()] {
		fmt.Println(v.Name)
		if v.Name == methodName {
			return v
		}
	}
	return nil
}

func (rmr RPCManager) TestMethod(v int, s string, b bool) int {
	fmt.Println("here is in TestMethod ",v, s)
	return 1
}
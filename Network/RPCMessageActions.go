package Network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/golang/protobuf/proto"
)

func (rm *RPCMessage) GetArgs() []reflect.Value {
	// -----
	fmt.Println("in get args :", len(rm.RPCRawArgs))
	args := make([]reflect.Value, len(rm.RPCRawArgs))
	for i := 0; i < len(rm.RPCRawArgs); i++ {
		args[i] = rm.RPCRawArgs[i].Value()
	}
	return args
}

func (rm *RPCMessage) SetArgs(args []interface{}) {
	for i := 0; i < len(args); i++ {
		rawArg := RPCRawArg{}
		fmt.Println("type is ", reflect.ValueOf(args[i]), reflect.TypeOf(args[i]))
		_, ok := args[i].(proto.Message)
		if ok {
			fmt.Println("success assert")
		} else {
			fmt.Println("failed assert")
		}
		rawArg.SetValue(args[i])
		rm.RPCRawArgs = append(rm.RPCRawArgs, &rawArg)
	}
}

func (rra *RPCRawArg) Value() reflect.Value {
	if rra.RawValue == nil {
		return reflect.ValueOf(nil)
	}
	switch rra.RawValueType {
	case RPCArgType_INT, RPCArgType_UInt:
		return reflect.ValueOf(binary.BigEndian.Uint32(rra.RawValue))
	case RPCArgType_Long, RPCArgType_ULong:
		return reflect.ValueOf(binary.BigEndian.Uint64(rra.RawValue))
	case RPCArgType_Short, RPCArgType_UShort:
		return reflect.ValueOf(binary.BigEndian.Uint16(rra.RawValue))
	case RPCArgType_Float:
		return reflect.ValueOf(math.Float32frombits(binary.BigEndian.Uint32(rra.RawValue)))
	case RPCArgType_Double:
		return reflect.ValueOf(math.Float64frombits(binary.BigEndian.Uint64(rra.RawValue)))
	case RPCArgType_String:
		return reflect.ValueOf(string(rra.RawValue))
	case RPCArgType_Bytes, RPCArgType_PBObject:
		return reflect.ValueOf(rra.RawValue)
	case RPCArgType_Bool:
		b, _ := strconv.ParseBool(string(rra.RawValue))
		return reflect.ValueOf(b)
	}
	return reflect.ValueOf(nil)
}

func (rra *RPCRawArg) SetValue(v interface{}) {
	fmt.Println("setvalue type is: ", reflect.TypeOf(reflect.ValueOf(v)))
	switch v.(type) {
	case int32:
		rra.RawValueType = RPCArgType_INT
		bytesBuffer := bytes.NewBuffer([]byte{})
		_ = binary.Read(bytesBuffer, binary.BigEndian, v.(int))
		rra.RawValue = bytesBuffer.Bytes()
	case string:
		rra.RawValueType = RPCArgType_String
		rra.RawValue = []byte(v.(string))
	case float32:
		rra.RawValueType = RPCArgType_Float
		bits := math.Float32bits(v.(float32))
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, bits)
		rra.RawValue = buf
	case float64:
		rra.RawValueType = RPCArgType_Double
		bits := math.Float64bits(v.(float64))
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, bits)
		rra.RawValue = buf
	case []byte:
		rra.RawValueType = RPCArgType_Bytes
		rra.RawValue = v.([]byte)
	case proto.Message:
		fmt.Println("proto.message")
		rra.RawValueType = RPCArgType_PBObject
		rra.RawValue = SerializeRPCMsg(v.(proto.Message))
	}
}

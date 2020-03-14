package Network

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestRPCMessage_String(t *testing.T) {
	rm := RPCMessage{}
	rm.Name = "test"
	rra := RPCRawArg{}
	rra.RawValueType = RPCArgType_INT
	rra.RawValue = []byte{0,0,0,1}
	rm.RPCRawArgs = append(rm.RPCRawArgs, &rra)
	ans := rm.String()
	fmt.Println(ans,binary.BigEndian.Uint32(rra.RawValue))
	return
}


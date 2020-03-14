package Network

import (
	"fmt"
	"testing"
)

func TestRPCRawArg_SetValue(t *testing.T) {
	rra := RPCRawArg{}
	rra.SetValue("value")
	fmt.Println(rra.RawValue,rra.RawValueType)
	return
}

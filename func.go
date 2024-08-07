package gruby

// #include <stdlib.h>
// #include "gruby.h"
import "C"

import (
	"unsafe"
)

// Func is the signature of a function in Go that you use to expose to Ruby
// code.
//
// The first return value is the actual return value for the code.
//
// The second return value is an exception, if any. This will be raised.
type Func func(m *GRuby, self Value) (Value, Value)

//export goMRBFuncCall
func goMRBFuncCall(state *C.mrb_state, value C.mrb_value) C.mrb_value {
	grb := states.Get(state)
	// Get the call info, which we use to lookup the proc
	callInfo := state.c.ci

	// Lookup the class itself
	class := *(**C.struct_RClass)(unsafe.Pointer(&callInfo.u[0]))
	methodTable := grb.classes[class]
	if methodTable == nil {
		panic("func call on unknown class")
	}

	// Lookup the method
	method := methodTable[callInfo.mid]
	if method == nil {
		panic("func call on unknown method")
	}

	result, exc := method(grb, grb.value(value))

	if result == nil {
		result = grb.NilValue()
	}

	if exc != nil {
		state.exc = C._go_mrb_getobj(exc.CValue())
		return grb.NilValue().CValue()
	}

	return result.CValue()
}

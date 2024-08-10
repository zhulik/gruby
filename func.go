package gruby

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
type Func func(grb *GRuby, self Value) (Value, Value)

//export goGRBFuncCall
func goGRBFuncCall(state *C.mrb_state, value C.mrb_value) C.mrb_value {
	grb := states.get(state)
	// Get the call info, which we use to lookup the proc
	callInfo := state.c.ci

	// Lookup the class itself
	class := *(**C.struct_RClass)(unsafe.Pointer(&callInfo.u[0]))

	method := grb.methods.get(class, callInfo.mid)

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

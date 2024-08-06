package gruby

// #include <stdlib.h>
// #include "gruby.h"
import "C"

import (
	"fmt"
	"sync"
	"unsafe"
)

// Func is the signature of a function in Go that you use to expose to Ruby
// code.
//
// The first return value is the actual return value for the code.
//
// The second return value is an exception, if any. This will be raised.
type Func func(m *GRuby, self Value) (Value, Value)

type (
	classMethodMap map[*C.struct_RClass]*methods
	methodMap      map[C.mrb_sym]Func
	stateMethodMap map[*C.mrb_state]*classMethods
)

type classMethods struct {
	Map   classMethodMap
	Mutex *sync.Mutex
}

type methods struct {
	Map   methodMap
	Mutex *sync.Mutex
}

type stateMethods struct {
	Map   stateMethodMap
	Mutex *sync.Mutex
}

// stateMethodTable is the lookup table for methods that we define in Go and
// expose in Ruby. This is cleaned up by Mrb.Close.
var stateMethodTable = &stateMethods{ //nolint:gochecknoglobals
	Mutex: new(sync.Mutex),
	Map:   make(stateMethodMap),
}

//export goMRBFuncCall
func goMRBFuncCall(state *C.mrb_state, value C.mrb_value) C.mrb_value {
	// Lookup the classes that we've registered methods for in this state
	stateMethodTable.Mutex.Lock()
	classTable := stateMethodTable.Map[state]
	stateMethodTable.Mutex.Unlock()
	if classTable == nil {
		panic(fmt.Sprintf("func call from unknown state: %p", state))
	}

	// Get the call info, which we use to lookup the proc
	callInfo := state.c.ci

	// Lookup the class itself
	classTable.Mutex.Lock()

	class := *(**C.struct_RClass)(unsafe.Pointer(&callInfo.u[0]))
	methodTable := classTable.Map[class]
	classTable.Mutex.Unlock()
	if methodTable == nil {
		panic("func call on unknown class")
	}

	// Lookup the method
	methodTable.Mutex.Lock()
	method := methodTable.Map[callInfo.mid]
	methodTable.Mutex.Unlock()
	if method == nil {
		panic("func call on unknown method")
	}

	// Call the method to get our *Value
	statesLock.RLock()
	mrb := states[state]
	statesLock.RUnlock()

	result, exc := method(mrb, mrb.value(value))

	if result == nil {
		result = mrb.NilValue()
	}

	if exc != nil {
		state.exc = C._go_mrb_getobj(exc.CValue())
		return mrb.NilValue().CValue()
	}

	return result.CValue()
}

func insertMethod(state *C.mrb_state, class *C.struct_RClass, name string, callback Func) {
	stateMethodTable.Mutex.Lock()
	classLookup := stateMethodTable.Map[state]
	if classLookup == nil {
		classLookup = &classMethods{Map: make(classMethodMap), Mutex: new(sync.Mutex)}
		stateMethodTable.Map[state] = classLookup
	}
	stateMethodTable.Mutex.Unlock()

	classLookup.Mutex.Lock()
	methodLookup := classLookup.Map[class]
	if methodLookup == nil {
		methodLookup = &methods{Map: make(methodMap), Mutex: new(sync.Mutex)}
		classLookup.Map[class] = methodLookup
	}
	classLookup.Mutex.Unlock()

	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	sym := C.mrb_intern_cstr(state, cstr)
	methodLookup.Mutex.Lock()
	methodLookup.Map[sym] = callback
	methodLookup.Mutex.Unlock()
}

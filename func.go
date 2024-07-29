package mruby

// #include <stdlib.h>
// #include "gomruby.h"
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
type Func func(m *Mrb, self Value) (Value, Value)

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
func goMRBFuncCall(s *C.mrb_state, v C.mrb_value) C.mrb_value {
	// Lookup the classes that we've registered methods for in this state
	stateMethodTable.Mutex.Lock()
	classTable := stateMethodTable.Map[s]
	stateMethodTable.Mutex.Unlock()
	if classTable == nil {
		panic(fmt.Sprintf("func call from unknown state: %p", s))
	}

	// Get the call info, which we use to lookup the proc
	ci := s.c.ci

	// Lookup the class itself
	classTable.Mutex.Lock()

	class := *(**C.struct_RClass)(unsafe.Pointer(&ci.u[0]))
	methodTable := classTable.Map[class]
	classTable.Mutex.Unlock()
	if methodTable == nil {
		panic("func call on unknown class")
	}

	// Lookup the method
	methodTable.Mutex.Lock()
	f := methodTable.Map[ci.mid]
	methodTable.Mutex.Unlock()
	if f == nil {
		panic("func call on unknown method")
	}

	// Call the method to get our *Value
	// TODO(mitchellh): reuse the Mrb instead of allocating every time
	mrb := &Mrb{s}
	result, exc := f(mrb, mrb.value(v))

	if result == nil {
		result = mrb.NilValue()
	}

	if exc != nil {
		s.exc = C._go_mrb_getobj(exc.CValue())
		return mrb.NilValue().CValue()
	}

	return result.CValue()
}

func insertMethod(s *C.mrb_state, c *C.struct_RClass, n string, f Func) {
	stateMethodTable.Mutex.Lock()
	classLookup := stateMethodTable.Map[s]
	if classLookup == nil {
		classLookup = &classMethods{Map: make(classMethodMap), Mutex: new(sync.Mutex)}
		stateMethodTable.Map[s] = classLookup
	}
	stateMethodTable.Mutex.Unlock()

	classLookup.Mutex.Lock()
	methodLookup := classLookup.Map[c]
	if methodLookup == nil {
		methodLookup = &methods{Map: make(methodMap), Mutex: new(sync.Mutex)}
		classLookup.Map[c] = methodLookup
	}
	classLookup.Mutex.Unlock()

	cs := C.CString(n)
	defer C.free(unsafe.Pointer(cs))

	sym := C.mrb_intern_cstr(s, cs)
	methodLookup.Mutex.Lock()
	methodLookup.Map[sym] = f
	methodLookup.Mutex.Unlock()
}

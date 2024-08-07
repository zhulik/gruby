package gruby

// #include "gruby.h"
import "C"

type (
	classMethodMap map[*C.struct_RClass]methodMap
	methodMap      map[C.mrb_sym]Func
)

type methodsStore struct {
	grb               *GRuby
	classes           classMethodMap
	getArgAccumulator []C.mrb_value
}

func (s *methodsStore) add(class *C.struct_RClass, name string, callback Func) {
	methodLookup := s.classes[class]
	if methodLookup == nil {
		methodLookup = make(methodMap)
		s.classes[class] = methodLookup
	}

	cstr := C.CString(name)
	defer freeStr(cstr)

	sym := C.mrb_intern_cstr(s.grb.state, cstr)
	methodLookup[sym] = callback
}

func (s *methodsStore) get(class *C.struct_RClass, name C.mrb_sym) Func {
	methodTable := s.grb.methods.classes[class]
	if methodTable == nil {
		panic("func call on unknown class")
	}

	// Lookup the method
	method := methodTable[name]
	if method == nil {
		panic("func call on unknown method")
	}

	return method
}

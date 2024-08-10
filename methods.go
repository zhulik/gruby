package gruby

// #include "gruby.h"
import "C"

type (
	classMethodMap map[*C.struct_RClass]methodMap
	methodMap      map[C.mrb_sym]Func
)

type methodsStore struct {
	grb     *GRuby
	classes classMethodMap
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
	// the caller _must_ call `add`` before calling `get`, crashes otherwise.
	return s.grb.methods.classes[class][name]
}

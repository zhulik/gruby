package gruby

import "unsafe"

// #include "gruby.h"
import "C"

// Class is a class in gruby. To obtain a Class, use DefineClass or
// one of the variants on the GRuby structure.
type Class struct {
	Value
	class *C.struct_RClass
}

// String returns the "to_s" result of this value.
func (c *Class) String() string {
	return MustToGo[string](c)
}

// DefineClassMethod defines a class-level method on the given class.
func (c *Class) DefineClassMethod(name string, cb Func, spec ArgSpec) {
	c.GRuby().insertClassMethod(c.class.c, name, cb)

	cstr := C.CString(name)
	defer freeStr(cstr)

	C.mrb_define_class_method(
		c.GRuby().state,
		c.class,
		cstr,
		C._go_grb_class_method_call_t(),
		C.mrb_aspec(spec))
}

// DefineConst defines a constant within this class.
func (c *Class) DefineConst(name string, value Value) {
	cstr := C.CString(name)
	defer freeStr(cstr)

	C.mrb_define_const(c.GRuby().state, c.class, cstr, value.CValue())
}

// DefineMethod defines an instance method on the class.
func (c *Class) DefineMethod(name string, cb Func, spec ArgSpec) {
	grb := c.GRuby()
	grb.insertInstanceMethod(c.class, name, cb)

	cstr := C.CString(name)
	defer freeStr(cstr)

	C.mrb_define_method(
		grb.state,
		c.class,
		cstr,
		C._go_grb_instance_method_call_t(),
		C.mrb_aspec(spec))
}

// New instantiates the class with the given args.
func (c *Class) New(args ...Value) (Value, error) {
	var argv []C.mrb_value
	var argvPtr *C.mrb_value
	if len(args) > 0 {
		// Make the raw byte slice to hold our arguments we'll pass to C
		argv = make([]C.mrb_value, len(args))
		for i, arg := range args {
			argv[i] = arg.CValue()
		}

		argvPtr = &argv[0]
	}

	result := C.mrb_obj_new(c.GRuby().state, c.class, C.mrb_int(len(argv)), argvPtr)
	if exc := checkException(c.GRuby()); exc != nil {
		return nil, exc
	}

	return c.GRuby().value(result), nil
}

func newClass(grb *GRuby, c *C.struct_RClass) *Class {
	return &Class{
		Value: grb.value(C.mrb_obj_value(unsafe.Pointer(c))),
		class: c,
	}
}

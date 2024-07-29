package mruby

import "unsafe"

// #include <stdlib.h>
// #include "gomruby.h"
import "C"

// Class is a class in mruby. To obtain a Class, use DefineClass or
// one of the variants on the Mrb structure.
type Class struct {
	Value
	class *C.struct_RClass
}

// String returns the "to_s" result of this value.
func (c *Class) String() string {
	return ToGo[string](c)
}

// DefineClassMethod defines a class-level method on the given class.
func (c *Class) DefineClassMethod(name string, cb Func, as ArgSpec) {
	insertMethod(c.Mrb().state, c.class.c, name, cb)

	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))

	C.mrb_define_class_method(
		c.Mrb().state,
		c.class,
		cs,
		C._go_mrb_func_t(),
		C.mrb_aspec(as))
}

// DefineConst defines a constant within this class.
func (c *Class) DefineConst(name string, value Value) {
	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))

	C.mrb_define_const(c.Mrb().state, c.class, cs, value.CValue())
}

// DefineMethod defines an instance method on the class.
func (c *Class) DefineMethod(name string, cb Func, as ArgSpec) {
	insertMethod(c.Mrb().state, c.class, name, cb)

	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))

	C.mrb_define_method(
		c.Mrb().state,
		c.class,
		cs,
		C._go_mrb_func_t(),
		C.mrb_aspec(as))
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

	result := C.mrb_obj_new(c.Mrb().state, c.class, C.mrb_int(len(argv)), argvPtr)
	if exc := checkException(c.Mrb().state); exc != nil {
		return nil, exc
	}

	return c.Mrb().value(result), nil
}

func newClass(mrb *Mrb, c *C.struct_RClass) *Class {
	return &Class{
		Value: mrb.value(C.mrb_obj_value(unsafe.Pointer(c))),
		class: c,
	}
}

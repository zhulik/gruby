package gruby

// #include "gruby.h"
import "C"

import (
	"errors"
	"unsafe"
)

var ErrEmptyArgs = errors.New("args must be non-empty and have a proc at the end")

// Value is an interface that should be implemented by anything that can
// be represents as an mruby value.
type Value interface { //nolint:interfacebloat
	String() string

	GRuby() *GRuby
	CValue() C.mrb_value

	Type() ValueType
	IsDead() bool
	Class() *Class
	SingletonClass() *Class

	SetInstanceVariable(variable string, value Value)
	GetInstanceVariable(variable string) Value

	Call(method string, args ...Value) (Value, error)
	CallBlock(method string, args ...Value) (Value, error)
}

// GValue is a "value" internally in gruby. A "value" is what mruby calls
// basically anything in Ruby: a class, an object (instance), a variable,
// etc.
type GValue struct {
	value C.mrb_value
	grb   *GRuby
}

// SetInstanceVariable sets an instance variable on this value.
func (v *GValue) SetInstanceVariable(variable string, value Value) {
	cstr := C.CString(variable)
	defer freeStr(cstr)
	C._go_mrb_iv_set(v.grb.state, v.value, C.mrb_intern_cstr(v.grb.state, cstr), value.CValue())
}

// GetInstanceVariable gets an instance variable on this value.
func (v *GValue) GetInstanceVariable(variable string) Value {
	cstr := C.CString(variable)
	defer freeStr(cstr)
	return v.grb.value(C._go_mrb_iv_get(v.grb.state, v.value, C.mrb_intern_cstr(v.grb.state, cstr)))
}

// Call calls a method with the given name and arguments on this
// value.
func (v *GValue) Call(method string, args ...Value) (Value, error) {
	return v.call(method, args, nil)
}

// CallBlock is the same as call except that it expects the last
// argument to be a Proc that will be passed into the function call.
// It is an error if args is empty or if there is no block on the end.
func (v *GValue) CallBlock(method string, args ...Value) (Value, error) {
	if len(args) == 0 {
		return nil, ErrEmptyArgs
	}

	n := len(args)
	return v.call(method, args[:n-1], args[n-1])
}

func (v *GValue) call(method string, args Values, block Value) (Value, error) {
	var argv []C.mrb_value
	var argvPtr *C.mrb_value

	if len(args) > 0 {
		// Make the raw byte slice to hold our arguments we'll pass to C*C.mrb
		argv = make([]C.mrb_value, len(args))
		for i, arg := range args {
			argv[i] = arg.CValue()
		}

		argvPtr = &argv[0]
	}

	var blockV *C.mrb_value
	if block != nil {
		val := block.CValue()
		blockV = &val
	}

	cstr := C.CString(method)
	defer freeStr(cstr)

	// If we have a block, we have to call a separate function to
	// pass a block in. Otherwise, we just call it directly.
	result := C._go_mrb_call(
		v.grb.state,
		v.value,
		C.mrb_intern_cstr(v.grb.state, cstr),
		C.mrb_int(len(argv)),
		argvPtr,
		blockV)

	if exc := checkException(v.grb); exc != nil {
		return nil, exc
	}

	return v.grb.value(result), nil
}

// IsDead tells you if an object has been collected by the GC or not.
func (v *GValue) IsDead() bool {
	return C.ushort(C._go_isdead(v.grb.state, v.value)) != 0
}

// GRuby returns the GRuby state for this value.
func (v *GValue) GRuby() *GRuby {
	return v.grb
}

// GCProtect protects this value from being garbage collected.
func (v *GValue) GCProtect() {
	C.mrb_gc_protect(v.grb.state, v.value)
}

// SetProcTargetClass sets the target class where a proc will be executed
// when this value is a proc.
func (v *GValue) SetProcTargetClass(c *Class) {
	proc := C._go_mrb_proc_ptr(v.value)

	*(**C.struct_RClass)(unsafe.Pointer(&proc.e[0])) = c.class
}

// Type returns the ValueType of the GValue. See the constants table.
func (v *GValue) Type() ValueType {
	if C._go_mrb_bool2int(C._go_mrb_nil_p(v.value)) == 1 {
		return TypeNil
	}

	return ValueType(C._go_mrb_type(v.value))
}

// CValue returns underlying mrb_value.
func (v *GValue) CValue() C.mrb_value {
	return v.value
}

// String returns the "to_s" result of this value.
func (v *GValue) String() string {
	return MustToGo[string](v)
}

// Class returns the *Class of a value.
func (v *GValue) Class() *Class {
	return newClass(v.grb, C.mrb_class(v.grb.state, v.value))
}

// SingletonClass returns the singleton class (a class isolated just for the
// scope of the object) for the given value.
func (v *GValue) SingletonClass() *Class {
	sclass := C._go_mrb_class_ptr(C.mrb_singleton_class(v.grb.state, v.value))
	return newClass(v.grb, sclass)
}

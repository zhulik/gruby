package mruby

// #include <stdlib.h>
// #include "gomruby.h"
import "C"

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

// Value is an interface that should be implemented by anything that can
// be represents as an mruby value.
type Value interface { //nolint:interfacebloat
	String() string

	Mrb() *Mrb
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

// MrbValue is a "value" internally in mruby. A "value" is what mruby calls
// basically anything in Ruby: a class, an object (instance), a variable,
// etc.
type MrbValue struct {
	value C.mrb_value
	state *C.mrb_state
}

// SetInstanceVariable sets an instance variable on this value.
func (v *MrbValue) SetInstanceVariable(variable string, value Value) {
	cs := C.CString(variable)
	defer C.free(unsafe.Pointer(cs))
	C._go_mrb_iv_set(v.state, v.value, C.mrb_intern_cstr(v.state, cs), value.CValue())
}

// GetInstanceVariable gets an instance variable on this value.
func (v *MrbValue) GetInstanceVariable(variable string) Value {
	cs := C.CString(variable)
	defer C.free(unsafe.Pointer(cs))
	return v.Mrb().value(C._go_mrb_iv_get(v.state, v.value, C.mrb_intern_cstr(v.state, cs)))
}

// Call calls a method with the given name and arguments on this
// value.
func (v *MrbValue) Call(method string, args ...Value) (Value, error) {
	return v.call(method, args, nil)
}

// CallBlock is the same as call except that it expects the last
// argument to be a Proc that will be passed into the function call.
// It is an error if args is empty or if there is no block on the end.
func (v *MrbValue) CallBlock(method string, args ...Value) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("args must be non-empty and have a proc at the end")
	}

	n := len(args)
	return v.call(method, args[:n-1], args[n-1])
}

func (v *MrbValue) call(method string, args []Value, block Value) (Value, error) {
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

	var blockV *C.mrb_value
	if block != nil {
		val := block.CValue()
		blockV = &val
	}

	cs := C.CString(method)
	defer C.free(unsafe.Pointer(cs))

	// If we have a block, we have to call a separate function to
	// pass a block in. Otherwise, we just call it directly.
	result := C._go_mrb_call(
		v.state,
		v.value,
		C.mrb_intern_cstr(v.state, cs),
		C.mrb_int(len(argv)),
		argvPtr,
		blockV)

	if exc := checkException(v.state); exc != nil {
		return nil, exc
	}

	return v.Mrb().value(result), nil
}

// IsDead tells you if an object has been collected by the GC or not.
func (v *MrbValue) IsDead() bool {
	return C.ushort(C._go_isdead(v.state, v.value)) != 0
}

// Mrb returns the Mrb state for this value.
func (v *MrbValue) Mrb() *Mrb {
	return &Mrb{v.state}
}

// GCProtect protects this value from being garbage collected.
func (v *MrbValue) GCProtect() {
	C.mrb_gc_protect(v.state, v.value)
}

// SetProcTargetClass sets the target class where a proc will be executed
// when this value is a proc.
func (v *MrbValue) SetProcTargetClass(c *Class) {
	proc := C._go_mrb_proc_ptr(v.value)

	*(**C.struct_RClass)(unsafe.Pointer(&proc.e[0])) = c.class
}

// Type returns the ValueType of the MrbValue. See the constants table.
func (v *MrbValue) Type() ValueType {
	if C._go_mrb_bool2int(C._go_mrb_nil_p(v.value)) == 1 {
		return TypeNil
	}

	return ValueType(C._go_mrb_type(v.value))
}

// CValue returns underlying mrb_value.
func (v *MrbValue) CValue() C.mrb_value {
	return v.value
}

// Exception is a special type of value that represents an error
// and implements the Error interface.
type Exception struct {
	Value
	File      string
	Line      int
	Message   string
	Backtrace []string
}

func (e *Exception) Error() string {
	return e.Message
}

//-------------------------------------------------------------------
// Type conversion to Go types
//-------------------------------------------------------------------

func ToGo[T any](value Value) T {
	var t T

	var result any

	switch any(t).(type) {
	case string:
		str := C.mrb_obj_as_string(value.Mrb().state, value.CValue())
		result = C.GoString(C._go_RSTRING_PTR(str))
	case int, int16, int32, int64:
		result = int(C._go_mrb_fixnum(value.CValue()))
	case float64, float32:
		result = float64(C._go_mrb_float(value.CValue()))
	case *Array:
		result = &Array{value}
	case *Hash:
		result = &Hash{value}
	default:
		panic(fmt.Sprintf("unknown type %+v", value))
	}

	return result.(T)
}

// String returns the "to_s" result of this value.
func (v *MrbValue) String() string {
	return ToGo[string](v)
}

// Class returns the *Class of a value.
func (v *MrbValue) Class() *Class {
	mrb := &Mrb{v.state}
	return newClass(mrb, C.mrb_class(v.state, v.value))
}

// SingletonClass returns the singleton class (a class isolated just for the
// scope of the object) for the given value.
func (v *MrbValue) SingletonClass() *Class {
	mrb := &Mrb{v.state}
	sclass := C._go_mrb_class_ptr(C.mrb_singleton_class(v.state, v.value))
	return newClass(mrb, sclass)
}

//-------------------------------------------------------------------
// Internal Functions
//-------------------------------------------------------------------

func newExceptionValue(s *C.mrb_state) *Exception {
	mrb := &Mrb{s}

	if s.exc == nil {
		panic("exception value init without exception")
	}

	arenaIndex := C._go_mrb_gc_arena_save(s)
	defer C._go_mrb_gc_arena_restore(s, arenaIndex)

	// Convert the RObject* to an mrb_value
	value := C.mrb_obj_value(unsafe.Pointer(s.exc))

	// Retrieve and convert backtrace to []string (avoiding reflection in Decode)
	var backtrace []string
	mrbBacktraceValue := mrb.value(C.mrb_exc_backtrace(s, value))
	if mrbBacktraceValue.Type() == TypeArray {
		mrbBacktrace := ToGo[*Array](mrbBacktraceValue)
		for i := 0; i < mrbBacktrace.Len(); i++ {
			ln, _ := mrbBacktrace.Get(i)
			backtrace = append(backtrace, ln.String())
		}
	}

	// Extract file + line from first backtrace line
	file := "Unknown"
	line := 0
	if len(backtrace) > 0 {
		fileAndLine := strings.Split(backtrace[0], ":")
		if len(fileAndLine) >= 2 {
			file = fileAndLine[0]
			line, _ = strconv.Atoi(fileAndLine[1])
		}
	}

	result := mrb.value(value)
	return &Exception{
		Value:     result,
		Message:   result.String(),
		File:      file,
		Line:      line,
		Backtrace: backtrace,
	}
}

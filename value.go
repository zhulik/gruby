package gruby

// #include <stdlib.h>
// #include "gruby.h"
import "C"

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

// MrbValue is a "value" internally in gruby. A "value" is what mruby calls
// basically anything in Ruby: a class, an object (instance), a variable,
// etc.
type MrbValue struct {
	value C.mrb_value
	grb   *GRuby
}

// SetInstanceVariable sets an instance variable on this value.
func (v *MrbValue) SetInstanceVariable(variable string, value Value) {
	cstr := C.CString(variable)
	defer C.free(unsafe.Pointer(cstr))
	C._go_mrb_iv_set(v.grb.state, v.value, C.mrb_intern_cstr(v.grb.state, cstr), value.CValue())
}

// GetInstanceVariable gets an instance variable on this value.
func (v *MrbValue) GetInstanceVariable(variable string) Value {
	cstr := C.CString(variable)
	defer C.free(unsafe.Pointer(cstr))
	return v.GRuby().value(C._go_mrb_iv_get(v.grb.state, v.value, C.mrb_intern_cstr(v.grb.state, cstr)))
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
		return nil, ErrEmptyArgs
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

	cstr := C.CString(method)
	defer C.free(unsafe.Pointer(cstr))

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

	return v.GRuby().value(result), nil
}

// IsDead tells you if an object has been collected by the GC or not.
func (v *MrbValue) IsDead() bool {
	return C.ushort(C._go_isdead(v.grb.state, v.value)) != 0
}

// GRuby returns the GRuby state for this value.
func (v *MrbValue) GRuby() *GRuby {
	return v.grb
}

// GCProtect protects this value from being garbage collected.
func (v *MrbValue) GCProtect() {
	C.mrb_gc_protect(v.grb.state, v.value)
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

// ExceptionError is a special type of value that represents an error
// and implements the Error interface.
type ExceptionError struct {
	Value
	File      string
	Line      int
	Message   string
	Backtrace []string
}

func (e *ExceptionError) Error() string {
	return e.Message
}

//-------------------------------------------------------------------
// Type conversion to Go types
//-------------------------------------------------------------------

func ToGo[T any](value Value) T {
	var empty T

	var result any

	switch any(empty).(type) {
	case string:
		str := C.mrb_obj_as_string(value.GRuby().state, value.CValue())
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

	return result.(T) //nolint:forcetypeassert
}

func ToRuby[T any](grb *GRuby, value T) Value {
	val := any(value)

	switch val.(type) {
	case bool:
		if val.(bool) {
			return grb.TrueValue()
		}
		return grb.FalseValue()
	case string:
		cstr := C.CString(val.(string)) //nolint:forcetypeassert
		defer C.free(unsafe.Pointer(cstr))
		return grb.value(C.mrb_str_new_cstr(grb.state, cstr))
	case int, int16, int32, int64:
		return grb.value(C.mrb_fixnum_value(C.mrb_int(val.(int)))) //nolint:forcetypeassert
	case float64, float32:
		return grb.value(C.mrb_float_value(grb.state, C.mrb_float(C.long(val.(float32))))) //nolint:forcetypeassert
	// TODO: generic array and hash support
	case []string:
		ary := NewArray(grb)

		for _, item := range val.([]string) {
			ary.Push(ToRuby(grb, item))
		}
		return ary
	}

	panic(fmt.Sprintf("unknown type '%+v'", value))
}

// String returns the "to_s" result of this value.
func (v *MrbValue) String() string {
	return ToGo[string](v)
}

// Class returns the *Class of a value.
func (v *MrbValue) Class() *Class {
	return newClass(v.grb, C.mrb_class(v.grb.state, v.value))
}

// SingletonClass returns the singleton class (a class isolated just for the
// scope of the object) for the given value.
func (v *MrbValue) SingletonClass() *Class {
	sclass := C._go_mrb_class_ptr(C.mrb_singleton_class(v.grb.state, v.value))
	return newClass(v.grb, sclass)
}

//-------------------------------------------------------------------
// Internal Functions
//-------------------------------------------------------------------

func newExceptionValue(grb *GRuby) *ExceptionError {
	state := grb.state
	if state.exc == nil {
		return nil
	}

	arenaIndex := C._go_mrb_gc_arena_save(state)
	defer C._go_mrb_gc_arena_restore(state, arenaIndex)

	// Convert the RObject* to an mrb_value
	value := C.mrb_obj_value(unsafe.Pointer(state.exc))

	// Retrieve and convert backtrace to []string (avoiding reflection in Decode)
	var backtrace []string
	mrbBacktraceValue := grb.value(C.mrb_exc_backtrace(state, value))
	if mrbBacktraceValue.Type() == TypeArray {
		mrbBacktrace := ToGo[*Array](mrbBacktraceValue)
		for i := range mrbBacktrace.Len() {
			ln, _ := mrbBacktrace.Get(i) //nolint:errcheck
			backtrace = append(backtrace, ln.String())
		}
	}

	// Extract file + line from first backtrace line
	file := "Unknown"
	line := 0
	if len(backtrace) > 0 {
		fileAndLine := strings.Split(backtrace[0], ":")
		if len(fileAndLine) >= 2 { //nolint:mnd
			file = fileAndLine[0]
			line, _ = strconv.Atoi(fileAndLine[1]) //nolint:errcheck
		}
	}

	result := grb.value(value)
	return &ExceptionError{
		Value:     result,
		Message:   result.String(),
		File:      file,
		Line:      line,
		Backtrace: backtrace,
	}
}

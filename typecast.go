package gruby

// #include "gruby.h"
import "C"

import "fmt"

// TODO: make sure all supported types covered in functions.
type SupportedTypes interface {
	bool | string | int | float32 | float64 | *Array | *Hash
}

func ToGo[T SupportedTypes](value Value) T {
	var empty T

	var result any

	switch any(empty).(type) {
	case string:
		str := C.mrb_obj_as_string(value.GRuby().state, value.CValue())
		result = C.GoString(C._go_RSTRING_PTR(str))
	case int:
		result = int(C._go_mrb_fixnum(value.CValue()))
	case float32:
		result = float32(C._go_mrb_float(value.CValue()))
	case float64:
		result = float64(C._go_mrb_float(value.CValue()))
	case *Array:
		result = &Array{value}
	case *Hash:
		result = &Hash{value}
	}

	return result.(T) //nolint:forcetypeassert
}

func ToRuby[T SupportedTypes](grb *GRuby, value T) Value {
	val := any(value)

	switch tVal := val.(type) {
	case bool:
		if tVal {
			return grb.TrueValue()
		}
		return grb.FalseValue()
	case string:
		cstr := C.CString(tVal)
		defer freeStr(cstr)
		return grb.value(C.mrb_str_new_cstr(grb.state, cstr))
	case int:
		return grb.value(C.mrb_fixnum_value(C.mrb_int(tVal)))
	case float32:
		return grb.value(C.mrb_float_value(grb.state, C.mrb_float(C.long(tVal))))
	case float64:
		return grb.value(C.mrb_float_value(grb.state, C.mrb_float(C.long(tVal))))
	// TODO: generic array and hash support
	case []string:
		ary := NewArray(grb)

		for _, item := range tVal {
			ary.Push(ToRuby(grb, item))
		}
		return ary
	}

	panic(fmt.Sprintf("unknown type '%+v'", value))
}

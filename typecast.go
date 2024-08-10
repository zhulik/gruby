package gruby

// #include "gruby.h"
import "C"

import (
	"fmt"
)

type (
	Values   []Value
	ValueMap map[Value]Value
)

type supportedComparables interface {
	comparable
	bool | string | int | float32 | float64
}

// TODO: make sure all supported types covered in functions.
type SupportedTypes interface {
	supportedComparables | Hash | Values
}

func ToGoArray[T SupportedTypes](array Values) []T {
	result := make([]T, len(array))

	for i, val := range array {
		result[i] = ToGo[T](val)
	}

	return result
}

func ToGoMap[K supportedComparables, V SupportedTypes](hash Hash) map[K]V {
	result := map[K]V{}

	keys := hash.Keys()

	for _, rbKey := range keys {
		key := ToGo[K](rbKey)
		val, err := hash.Get(rbKey)
		if err != nil {
			panic(err)
		}
		result[key] = ToGo[V](val)
	}

	return result
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
	case Hash:
		result = Hash{value}
	case Values:
		count := int(C._go_RARRAY_LEN(value.CValue()))
		goAry := make(Values, count)
		for i := range count {
			goAry[i] = value.GRuby().value(C.mrb_ary_entry(value.CValue(), C.mrb_int(i)))
		}
		result = goAry
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
	}

	panic(fmt.Sprintf("unknown type '%+v'", value))
}

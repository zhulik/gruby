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

type SupportedComparables interface {
	comparable
	bool | string | int | float32 | float64
}

// TODO: make sure all supported types covered in functions.
type SupportedTypes interface {
	SupportedComparables | Hash | Values
}

// TODO: Must version
func ToGoArray[T SupportedTypes](array Values) ([]T, error) {
	result := make([]T, len(array))

	for i, val := range array {
		gVal, err := ToGo[T](val)
		if err != nil {
			return nil, err
		}
		result[i] = gVal
	}

	return result, nil
}

func MustToGoArray[T SupportedTypes](array Values) []T {
	return Must(ToGoArray[T](array))
}

// TODO: Must version
func ToGoMap[K SupportedComparables, V SupportedTypes](hash Hash) (map[K]V, error) {
	result := map[K]V{}

	for _, rbKey := range hash.Keys() {
		key, err := ToGo[K](rbKey)
		if err != nil {
			return nil, err
		}
		v, err := ToGo[V](hash.Get(rbKey))
		if err != nil {
			return nil, err
		}
		result[key] = v
	}

	return result, nil
}

func ToGo[T SupportedTypes](value Value) (T, error) {
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
	default:
		return empty, fmt.Errorf("%w: '%+v'", ErrUnknownType, empty)
	}

	// We don't check the assertion because the value is created in this exact function,
	// just make sure it's correct.
	return result.(T), nil //nolint:forcetypeassert
}

func MustToGo[T SupportedTypes](value Value) T {
	return Must(ToGo[T](value))
}

func ToRuby[T SupportedTypes](grb *GRuby, value T) (Value, error) {
	val := any(value)

	switch tVal := val.(type) {
	case bool:
		if tVal {
			return grb.TrueValue(), nil
		}
		return grb.FalseValue(), nil
	case string:
		cstr := C.CString(tVal)
		defer freeStr(cstr)
		return grb.value(C.mrb_str_new_cstr(grb.state, cstr)), nil
	case int:
		return grb.value(C.mrb_fixnum_value(C.mrb_int(tVal))), nil
	case float32:
		return grb.value(C.mrb_float_value(grb.state, C.mrb_float(C.long(tVal)))), nil
	case float64:
		return grb.value(C.mrb_float_value(grb.state, C.mrb_float(C.long(tVal)))), nil
	}

	return nil, fmt.Errorf("%w: '%+v'", ErrUnknownType, value)
}

func MustToRuby[T SupportedTypes](grb *GRuby, value T) Value {
	return Must(ToRuby[T](grb, value))
}

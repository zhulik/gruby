package gruby

// #include "gruby.h"
import "C"

// Array represents an GValue that is a Array in Ruby.
//
// A Array can be obtained by calling the Array function on GValue.
type Array struct {
	Value
}

func NewArray(grb *GRuby) *Array {
	// TODO: who deletes it?
	return &Array{grb.value(C.mrb_ary_new(grb.state))}
}

// Push adds an item to the arrayss
func (v *Array) Push(item Value) {
	C.mrb_ary_push(v.GRuby().state, v.CValue(), item.CValue())
}

// Len returns the length of the array.
func (v *Array) Len() int {
	return int(C._go_RARRAY_LEN(v.CValue()))
}

// Get gets an element form the Array by index.
//
// This does not copy the element. This is a pointer/reference directly
// to the element in the array.
func (v *Array) Get(idx int) (Value, error) {
	result := C.mrb_ary_entry(v.CValue(), C.mrb_int(idx))

	val := v.GRuby().value(result)
	if val.Type() == TypeNil {
		val = nil
	}

	return val, nil
}

package mruby

// #include "gomruby.h"
import "C"

// Hash represents an MrbValue that is a Hash in Ruby.
//
// A Hash can be obtained by calling the Hash function on MrbValue.
type Hash struct {
	Value
}

// Delete deletes a key from the hash, returning its existing value,
// or nil if there wasn't a value.
func (h *Hash) Delete(key Value) Value {
	keyVal := key.CValue()
	result := C.mrb_hash_delete_key(h.Mrb().state, h.CValue(), keyVal)

	val := h.Mrb().value(result)
	if val.Type() == TypeNil {
		return nil
	}

	return val
}

// Get reads a value from the hash.
func (h *Hash) Get(key Value) (Value, error) {
	keyVal := key.CValue()
	result := C.mrb_hash_get(h.Mrb().state, h.CValue(), keyVal)
	return h.Mrb().value(result), nil
}

// Set sets a value on the hash
func (h *Hash) Set(key, val Value) error {
	keyVal := key.CValue()
	valVal := val.CValue()
	C.mrb_hash_set(h.Mrb().state, h.CValue(), keyVal, valVal)
	return nil
}

// Keys returns the array of keys that the Hash has. This is returned
// as an c since this is a Ruby array. You can iterate over it as
// you see fit.
func (h *Hash) Keys() (Value, error) {
	result := C.mrb_hash_keys(h.Mrb().state, h.CValue())
	return h.Mrb().value(result), nil
}

package gruby

// #include "gruby.h"
import "C"

// Hash represents an GValue that is a Hash in Ruby.
//
// A Hash can be obtained by calling the Hash function on GValue.
type Hash struct {
	Value
}

// Delete deletes a key from the hash, returning its existing value,
// or nil if there wasn't a value.
func (h *Hash) Delete(key Value) Value {
	result := C.mrb_hash_delete_key(h.GRuby().state, h.CValue(), key.CValue())

	val := h.GRuby().value(result)
	if val.Type() == TypeNil {
		return nil
	}

	return val
}

// Get reads a value from the hash.
func (h *Hash) Get(key Value) Value {
	result := C.mrb_hash_get(h.GRuby().state, h.CValue(), key.CValue())

	val := h.GRuby().value(result)
	if val.Type() == TypeNil {
		return nil
	}
	return h.GRuby().value(result)
}

// Set sets a value on the hash
func (h *Hash) Set(key, val Value) {
	C.mrb_hash_set(h.GRuby().state, h.CValue(), key.CValue(), val.CValue())
}

// Keys returns the array of keys that the Hash has as Values.
func (h *Hash) Keys() Values {
	keys := C.mrb_hash_keys(h.GRuby().state, h.CValue())

	return ToGo[Values](h.GRuby().value(keys))
}

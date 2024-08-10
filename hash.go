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
	keyVal := key.CValue()
	result := C.mrb_hash_delete_key(h.GRuby().state, h.CValue(), keyVal)

	val := h.GRuby().value(result)
	if val.Type() == TypeNil {
		return nil
	}

	return val
}

// Get reads a value from the hash.
func (h *Hash) Get(key Value) (Value, error) {
	keyVal := key.CValue()
	result := C.mrb_hash_get(h.GRuby().state, h.CValue(), keyVal)
	return h.GRuby().value(result), nil
}

// Set sets a value on the hash
func (h *Hash) Set(key, val Value) error {
	keyVal := key.CValue()
	valVal := val.CValue()
	C.mrb_hash_set(h.GRuby().state, h.CValue(), keyVal, valVal)
	return nil
}

// Keys returns the array of keys that the Hash has. This is returned
// as an c since this is a Ruby array. You can iterate over it as
// you see fit.
// TODO: return Values instead
func (h *Hash) Keys() (Values, error) {
	result := C.mrb_hash_keys(h.GRuby().state, h.CValue())

	keys := ToGo[Values](h.GRuby().value(result))
	return keys, nil
}

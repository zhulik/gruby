package gruby

// #include "gruby.h"
import "C"

import (
	"unsafe"

	"github.com/cornelk/hashmap"
)

var states = stateStore{ //nolint:gochecknoglobals
	states: hashmap.New[uintptr, *GRuby](),
}

type stateStore struct {
	states *hashmap.Map[uintptr, *GRuby]
}

func (s *stateStore) add(grb *GRuby) {
	s.states.Set(s.ptr(grb.state), grb)
}

func (s *stateStore) delete(grb *GRuby) {
	s.states.Del(s.ptr(grb.state))
}

func (s *stateStore) get(state *C.mrb_state) *GRuby {
	// the caller _must_ call `add`` before calling `get`, returns nil otherwise.
	grb, _ := s.states.Get(s.ptr(state))
	return grb
}

func (s *stateStore) ptr(state *C.mrb_state) uintptr {
	return uintptr(unsafe.Pointer(state))
}

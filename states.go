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

func (s *stateStore) Add(grb *GRuby) {
	s.states.Set(uintptr(unsafe.Pointer(grb.state)), grb)
}

func (s *stateStore) Delete(grb *GRuby) {
	s.states.Del(uintptr(unsafe.Pointer(grb.state)))
}

func (s *stateStore) Get(state *C.mrb_state) *GRuby {
	grb, ok := s.states.Get(uintptr(unsafe.Pointer(state)))
	if !ok {
		panic("state not found, this must never happen")
	}
	return grb
}

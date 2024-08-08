package gruby

// #include "gruby.h"
import "C"

import (
	"fmt"
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
	s.states.Set(s.ptr(grb.state), grb)
}

func (s *stateStore) Delete(grb *GRuby) {
	s.states.Del(s.ptr(grb.state))
}

func (s *stateStore) Get(state *C.mrb_state) *GRuby {
	grb, ok := s.states.Get(s.ptr(state))
	if !ok {
		panic(fmt.Sprintf("state not found, this must never happen: %d", s.ptr(state)))
	}
	return grb
}

func (s *stateStore) ptr(state *C.mrb_state) uintptr {
	return uintptr(unsafe.Pointer(state))
}

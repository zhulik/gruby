package gruby

import "sync"

// #include "gruby.h"
import "C"

var states = stateStore{ //nolint:gochecknoglobals
	states: map[*C.mrb_state]*GRuby{},
	lock:   sync.RWMutex{},
}

type stateStore struct {
	states map[*C.mrb_state]*GRuby
	lock   sync.RWMutex
}

func (s *stateStore) Add(state *C.mrb_state, grb *GRuby) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.states[state] = grb
}

func (s *stateStore) Delete(state *C.mrb_state) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.states, state)
}

func (s *stateStore) Get(state *C.mrb_state) *GRuby {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.states[state]
}

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

func (s *stateStore) Add(grb *GRuby) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.states[grb.state] = grb
}

func (s *stateStore) Delete(grb *GRuby) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.states, grb.state)
}

func (s *stateStore) Get(state *C.mrb_state) *GRuby {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.states[state]
}

package state

import "sync"

type Atom struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

type AlsetState struct {
	mu    sync.RWMutex
	Atoms map[string]*Atom `json:"atoms"`
}

func New() *AlsetState {
	return &AlsetState{Atoms: map[string]*Atom{}}
}

func (s *AlsetState) Get(id string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if a, ok := s.Atoms[id]; ok {
		return a.Value
	}
	return nil
}

func (s *AlsetState) Set(id string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Atoms[id] = &Atom{ID: id, Value: v}
}

func (s *AlsetState) Snapshot() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := map[string]interface{}{}
	for k, a := range s.Atoms {
		out[k] = a.Value
	}
	return out
}

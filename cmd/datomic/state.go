package main

import (
	"slices"
	"sync"
)

type State struct {
	Data map[int][]int
	Lock sync.Mutex
}

func NewState() State {
	return State{
		Data: map[int][]int{},
	}
}

func (s *State) Transact(transactions []*Transaction) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	for _, t := range transactions {
		if t.Op == "r" {
			t.Values = slices.Clone(s.Data[t.Key])
		}

		if t.Op == "append" {
			s.Data[t.Key] = append(s.Data[t.Key], t.Values...)
		}
	}
}

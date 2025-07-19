package main

import "sync"

type setItem interface {
	int
}

type set[T setItem] struct {
	mutex *sync.RWMutex
	m     map[T]any
}

func newSet[T setItem]() set[T] {
	return set[T]{
		mutex: &sync.RWMutex{},
		m:     map[T]any{},
	}
}

func (s *set[T]) union(newSet set[T]) {
	for newSetItem := range newSet.m {
		s.add(newSetItem)
	}
}

func (s *set[T]) serialize() []T {
	result := []T{}
	s.mutex.Lock()
	for item := range s.m {
		result = append(result, item)
	}
	s.mutex.Unlock()

	return result
}

func deserialize[T setItem](items []T) set[T] {
	newSet := newSet[T]()

	for _, item := range items {
		newSet.add(item)
	}

	return newSet
}

func (s *set[T]) add(element T) {
	s.mutex.Lock()
	s.m[element] = true
	s.mutex.Unlock()
}

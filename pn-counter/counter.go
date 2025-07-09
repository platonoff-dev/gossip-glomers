package main

import (
	"maps"
	"sync"
)

type pncounter struct {
	mutex     *sync.Mutex
	pCounters map[string]int
	nCounters map[string]int
}

func newCounter() pncounter {
	return pncounter{
		mutex:     &sync.Mutex{},
		pCounters: map[string]int{},
		nCounters: map[string]int{},
	}
}

func (c *pncounter) add(nodeID string, delta int) {
	c.mutex.Lock()

	if delta > 0 {
		counter := c.pCounters[nodeID]
		c.pCounters[nodeID] = counter + delta
	} else {
		counter := c.nCounters[nodeID]
		c.nCounters[nodeID] = counter + delta
	}

	c.mutex.Unlock()
}

func (c *pncounter) read() int {
	s := 0
	c.mutex.Lock()

	for _, v := range c.pCounters {
		s += v
	}

	for _, v := range c.nCounters {
		s += v
	}

	c.mutex.Unlock()
	return s
}

func (c *pncounter) merge(externalCounter pncounter) {
	c.mutex.Lock()

	for node, counter := range externalCounter.pCounters {
		c.pCounters[node] = max(c.pCounters[node], counter)
	}

	for node, counter := range externalCounter.nCounters {
		c.nCounters[node] = min(c.nCounters[node], counter)
	}

	c.mutex.Unlock()
}

type counter = map[string]int

func (c *pncounter) serialize() any {
	c.mutex.Lock()

	result := map[string]counter{
		"p": maps.Clone(c.pCounters),
		"n": maps.Clone(c.nCounters),
	}

	c.mutex.Unlock()
	return result
}

func deserialize(data any) pncounter {
	c := newCounter()

	counters := data.(map[string]counter)
	c.pCounters = counters["p"]
	c.nCounters = counters["n"]
	return c
}

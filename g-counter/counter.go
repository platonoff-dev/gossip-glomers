package main

import (
	"sync"
)

type gcounter struct {
	mutex        *sync.Mutex
	nodeCounters map[string]int
}

func newCounter() gcounter {
	return gcounter{
		mutex:        &sync.Mutex{},
		nodeCounters: map[string]int{},
	}
}

func (c *gcounter) add(nodeID string, delta int) {
	c.mutex.Lock()

	counter := c.nodeCounters[nodeID]
	c.nodeCounters[nodeID] = counter + delta

	c.mutex.Unlock()
}

func (c *gcounter) read() int {
	s := 0
	c.mutex.Lock()
	for _, v := range c.nodeCounters {
		s += v
	}

	c.mutex.Unlock()
	return s
}

func (c *gcounter) merge(externalCounter gcounter) {
	c.mutex.Lock()

	for node, counter := range externalCounter.nodeCounters {
		if counter > c.nodeCounters[node] {
			c.nodeCounters[node] = counter
		}
	}

	c.mutex.Unlock()
}

func (c *gcounter) serialize() any {
	return c.nodeCounters
}

func deserialize(data any) gcounter {
	counter := newCounter()
	counter.nodeCounters = data.(map[string]int)
	return counter
}

package callcounter

import (
	"sync"
)

type CallCounterRWMutex struct {
	sync.RWMutex
	counters map[string]int64
}

func NewCallCounterRWMutex() *CallCounterRWMutex {
	c := CallCounterRWMutex{
		counters: make(map[string]int64),
	}

	return &c
}

func (c *CallCounterRWMutex) Call(name string) {
	c.Lock()
	c.counters[name]++
	c.Unlock()
}

func (c *CallCounterRWMutex) Count(name string) int64 {
	c.RLock()
	count := c.counters[name]
	c.RUnlock()

	return count
}

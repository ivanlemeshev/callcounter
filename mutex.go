package callcounter

import (
	"sync"
)

type CallCounterMutex struct {
	sync.Mutex
	counters map[string]int64
}

func NewCallCounterMutex() *CallCounterMutex {
	c := CallCounterMutex{
		counters: make(map[string]int64),
	}

	return &c
}

func (c *CallCounterMutex) Call(name string) {
	c.Lock()
	c.counters[name]++
	c.Unlock()
}

func (c *CallCounterMutex) Count(name string) int64 {
	c.Lock()
	count := c.counters[name]
	c.Unlock()

	return count
}

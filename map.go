package callcounter

import (
	"sync"
	"sync/atomic"
)

type CallCounterMap struct {
	sync.Map
}

func NewCallCounterMap() *CallCounterMap {
	c := CallCounterMap{}

	return &c
}

func (c *CallCounterMap) Call(name string) {
	value := int64(1)
	count, loaded := c.LoadOrStore(name, &value)
	if loaded {
		atomic.AddInt64(count.(*int64), int64(value))
	}
}

func (s *CallCounterMap) Count(name string) int64 {
	count, ok := s.Load(name)
	if ok {
		return atomic.LoadInt64(count.(*int64))
	}

	return 0
}

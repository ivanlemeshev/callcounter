package callcounter

type CallCounterChannelMutex struct {
	counters map[string]int64
	mutex    chan struct{}
}

func NewCallCounterChannelMutex() *CallCounterChannelMutex {
	c := CallCounterChannelMutex{
		counters: make(map[string]int64),
		mutex:    make(chan struct{}, 1),
	}

	return &c
}

func (c *CallCounterChannelMutex) Call(name string) {
	c.mutex <- struct{}{}
	c.counters[name]++
	<-c.mutex
}

func (c *CallCounterChannelMutex) Count(name string) int64 {
	c.mutex <- struct{}{}
	count := c.counters[name]
	<-c.mutex
	return count
}

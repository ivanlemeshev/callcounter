package callcounter

import "context"

type CallCounterChannels struct {
	counters map[string]int64
	call     chan string
	count    chan string
	result   chan int64
}

func NewCallCounterChannels(ctx context.Context) *CallCounterChannels {
	c := CallCounterChannels{
		counters: make(map[string]int64),
		call:     make(chan string),
		count:    make(chan string),
		result:   make(chan int64),
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case name := <-c.call:
				c.counters[name]++
			case name := <-c.count:
				c.result <- c.counters[name]
			}
		}
	}()

	return &c
}

func (c *CallCounterChannels) Call(name string) {
	c.call <- name
}

func (c *CallCounterChannels) Count(name string) int64 {
	c.count <- name

	return <-c.result
}

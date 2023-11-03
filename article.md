# Different ways to protect maps from concurrent access in Golang.

## The problem

Periodically, when you are developing a Golang application, there is a need to store some data in memory using a map data structure. At the same time, the application can concurrently process a large number of requests and use the map in several goroutines concurrently. If you do not use any mechanisms of synchronization, the code fails with panic in runtime in that case. It is important to think of concurrent access to the map and data races. In this article, I want to consider several ways to solve this problem. Golang has built-in synchronization primitives and channels that can be used for that. Depending on the specific case, you can use different approaches.

I am going to create the call counter that is a simple component that counts the number of times a method is called and the counters will be stored in a map. To begin with, let's define a simple interface `CallCounter` that will have two methods. The method `Call` will update the counter when the method is called, and the `Count` method will return the current value for the given method name.

```go
package callcounter

type CallCounter interface {
	Call(name string)
	Count(name string) int64
}

```

Now let's consider different implementations.

## sync.Mutex

`sync.Mutex` is a mutual exclusion lock. It is a synchronization primitive that allows concurrent goroutines to safely access shared data. A mutex ensures that only one goroutine can access the shared data at a time, preventing data races and other concurrency problems.

To use a `sync.Mutex`, you must first lock it before accessing the shared data. Once you are finished accessing the shared data, you must unlock the mutex. If another goroutine tries to lock the mutex while it is already locked, the goroutine will be blocked until the mutex is unlocked.

Here is an implementation using `sync.Mutex`:

```go
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
```

## sync.RWMutex

`sync.RWMutex` is a read-write mutual exclusion lock. It is a synchronization primitive that allows multiple goroutines to concurrently read shared data, but only one goroutine to write shared data at a time. This is useful for situations where you need to allow concurrent reads, but not concurrent writes, to shared data.

To use a `sync.RWMutex`, you must first lock it for reading and writing using the `RLock` and `Lock` methods, respectively. Once you are finished accessing the shared data, you must unlock the mutex using the `RUnlock` or `Unlock` methods, respectively.

If another goroutine tries to lock the mutex for reading while it is locked for writing, the goroutine will be blocked until the mutex is unlocked. Similarly, if another goroutine tries to lock the mutex for writing while it is locked for reading, the goroutine will be blocked until the mutex is unlocked.

Here is an implementation using `sync.RWMutex`:

```go
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
```

## sync.Map

`sync.Map` is a concurrent map implementation in Golang. It is a specialized map implementation that provides a concurrent, thread-safe map. It is specifically designed for use cases where the entry set of the map is stable over time, such as caches.

`sync.Map` is a safe map, meaning that it can be accessed by multiple goroutines concurrently without the risk of race conditions or data corruption. It is also a fast map, with amortized constant-time loads, stores, and deletes.

Here is an implementation using `sync.Map`:

```go
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
```

## Channels

Channels are a powerful tool for communicating and synchronizing between concurrent goroutines. They are typed conduits through which you can send and receive values.

To create a channel, you use the `make` function. The `make` function takes two arguments: the type of the values that the channel will transmit and the capacity of the channel. The capacity of the channel is the maximum number of values that the channel can store without blocking the sender.

Channels allow us to write the code in a way where only one goroutine reads or writes to the map.

Here is an implementation using channels:

```go
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
```

## Mutex based on a channel

Also, Golang allows us to implement a simple mutex using channels. We need to use a buffered channel for that. A buffered channel in Golang is a channel that can store a limited number of values without blocking the sender. This is in contrast to an unbuffered channel, which can only store one value at a time and will block the sender if there is no receiver ready to receive the value.

Buffered channels are created by passing an additional capacity parameter to the `make` function. This parameter specifies the maximum number of values that the channel can store. We need to create a buffered channel that can store exactly one value. It will allow us to use this channel to lock and unlock goroutines because when the buffered channel is full it blocks other goroutins from writing into it and they will wait until it is free.

Here is an implementation using a buffered channel:

```go
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
```

## Comparing and testing

Now let's write a simple test to ensure that our code is correct, concurrent safe, and does not have data races. We will also write some benchmarks and compare the performance of different approaches.

```go
package callcounter_test

import (
	"callcounter"
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallCounter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tt := []struct {
		name    string
		counter callcounter.CallCounter
		calls   []int
	}{
		{
			name:    "mutex",
			counter: callcounter.NewCallCounterMutex(),
			calls:   []int{0, 1, 3, 100},
		},
		{
			name:    "rwmutex",
			counter: callcounter.NewCallCounterRWMutex(),
			calls:   []int{0, 1, 3, 100},
		},
		{
			name:    "map",
			counter: callcounter.NewCallCounterMap(),
			calls:   []int{0, 1, 3, 100},
		},
		{
			name:    "channels",
			counter: callcounter.NewCallCounterChannels(ctx),
			calls:   []int{0, 1, 3, 100},
		},
		{
			name:    "mutex based on a channel",
			counter: callcounter.NewCallCounterChannelMutex(),
			calls:   []int{0, 1, 3, 100},
		},
	}

	for _, tc := range tt {
		for _, calls := range tc.calls {
			testName := fmt.Sprintf("%s %d calls", tc.name, calls)

			t.Run(testName, func(t *testing.T) {
				wg := sync.WaitGroup{}
				for i := 0; i < calls; i++ {
					wg.Add(1)

					// Run in goroutine to make sure
					// there is no any issue with concurrent counter.
					go func() {
						defer wg.Done()

						tc.counter.Call(testName)
					}()
				}

				wg.Wait()
				assert.Equal(t, int64(calls), tc.counter.Count(testName))
			})
		}
	}
}

var result int64

func BenchmarkCallCounter(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tt := []struct {
		name    string
		counter callcounter.CallCounter
	}{
		{
			name:    "mutex",
			counter: callcounter.NewCallCounterMutex(),
		},
		{
			name:    "rwmutex",
			counter: callcounter.NewCallCounterRWMutex(),
		},
		{
			name:    "map",
			counter: callcounter.NewCallCounterMap(),
		},
		{
			name:    "channels",
			counter: callcounter.NewCallCounterChannels(ctx),
		},
		{
			name:    "mutex based on a channel",
			counter: callcounter.NewCallCounterChannelMutex(),
		},
	}

	var count int64

	for _, tc := range tt {
		b.Run(tc.name, func(b *testing.B) {
			wg := sync.WaitGroup{}
			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					tc.counter.Call("call_name")
				}()

				wg.Add(1)
				go func() {
					defer wg.Done()

					// Always record the result of Count to prevent
					// the compiler eliminating the function call.
					count = tc.counter.Count("call_name")
				}()
			}

			wg.Wait()

			// Always store the result to a package level variable
			// so the compiler cannot eliminate the Benchmark itself.
			result = count
		})
	}
}

```

Let's run the tests using the flag `--race` and benchmark tests:

```
go test --race -v .
=== RUN   TestCallCounter
=== RUN   TestCallCounter/mutex_0_calls
=== RUN   TestCallCounter/mutex_1_calls
=== RUN   TestCallCounter/mutex_3_calls
=== RUN   TestCallCounter/mutex_100_calls
=== RUN   TestCallCounter/rwmutex_0_calls
=== RUN   TestCallCounter/rwmutex_1_calls
=== RUN   TestCallCounter/rwmutex_3_calls
=== RUN   TestCallCounter/rwmutex_100_calls
=== RUN   TestCallCounter/map_0_calls
=== RUN   TestCallCounter/map_1_calls
=== RUN   TestCallCounter/map_3_calls
=== RUN   TestCallCounter/map_100_calls
=== RUN   TestCallCounter/channels_0_calls
=== RUN   TestCallCounter/channels_1_calls
=== RUN   TestCallCounter/channels_3_calls
=== RUN   TestCallCounter/channels_100_calls
=== RUN   TestCallCounter/mutex_channel_0_calls
=== RUN   TestCallCounter/mutex_channel_1_calls
=== RUN   TestCallCounter/mutex_channel_3_calls
=== RUN   TestCallCounter/mutex_channel_100_calls
--- PASS: TestCallCounter (0.01s)
    --- PASS: TestCallCounter/mutex_0_calls (0.00s)
    --- PASS: TestCallCounter/mutex_1_calls (0.00s)
    --- PASS: TestCallCounter/mutex_3_calls (0.00s)
    --- PASS: TestCallCounter/mutex_100_calls (0.00s)
    --- PASS: TestCallCounter/rwmutex_0_calls (0.00s)
    --- PASS: TestCallCounter/rwmutex_1_calls (0.00s)
    --- PASS: TestCallCounter/rwmutex_3_calls (0.00s)
    --- PASS: TestCallCounter/rwmutex_100_calls (0.00s)
    --- PASS: TestCallCounter/map_0_calls (0.00s)
    --- PASS: TestCallCounter/map_1_calls (0.00s)
    --- PASS: TestCallCounter/map_3_calls (0.00s)
    --- PASS: TestCallCounter/map_100_calls (0.00s)
    --- PASS: TestCallCounter/channels_0_calls (0.00s)
    --- PASS: TestCallCounter/channels_1_calls (0.00s)
    --- PASS: TestCallCounter/channels_3_calls (0.00s)
    --- PASS: TestCallCounter/channels_100_calls (0.00s)
    --- PASS: TestCallCounter/mutex_channel_0_calls (0.00s)
    --- PASS: TestCallCounter/mutex_channel_1_calls (0.00s)
    --- PASS: TestCallCounter/mutex_channel_3_calls (0.00s)
    --- PASS: TestCallCounter/mutex_channel_100_calls (0.00s)
PASS
ok      github.com/ivanlemeshev/callcounter     1.014s
```

```
go test -bench=. -benchmem -count 1
goos: linux
goarch: amd64
pkg: github.com/ivanlemeshev/callcounter
cpu: AMD Ryzen 7 5800H with Radeon Graphics
BenchmarkCallCounter/mutex-16            2054628               605.1 ns/op            79 B/op          2 allocs/op
BenchmarkCallCounter/rwmutex-16          1000000              1057 ns/op             174 B/op          2 allocs/op
BenchmarkCallCounter/map-16              3605562               363.4 ns/op            80 B/op          4 allocs/op
BenchmarkCallCounter/channels-16          593912              2483 ns/op             273 B/op          3 allocs/op
BenchmarkCallCounter/mutex_channel-16     492544              2310 ns/op             268 B/op          3 allocs/op
PASS
ok      github.com/ivanlemeshev/callcounter     7.923s
```

The benchmark results show that the Mutex-based implementation is the fastest and most efficient implementation. It has the lowest operation time and the lowest memory allocation. The RWMutex-based implementation is a good alternative if you need to support concurrent reads and writes to the call counter. The map-based implementation is a good alternative if you need a fast call counter implementation and you don't need to support concurrent writes. The channel-based implementation is the slowest and least efficient implementation in our case.

## Summary

The results show that depending on the requirements you should choose the approach that is most suitable for your particular case. It may be a bad idea to use only channels all the time.

`sync.Mutex`, `sync.RWMutex`, `sync.Map` and channels are synchronization primitives in Golang that can be used to protect shared data and coordinate the execution of goroutines. However, they have different strengths and weaknesses and should be used in different situations.

Mutexes are a good choice for protecting shared data that is frequently accessed by concurrent goroutines. For example, you might use a mutex to protect a shared counter or a map.

Channels are a good choice for coordinating the execution of goroutines that need to communicate with each other or synchronize their execution. For example, you might use channels to implement a producer-consumer pattern or to coordinate the execution of a pipeline of goroutines.

`sync.Map` can be used when you need a concurrent-safe map that is optimized for read-heavy workloads. `sync.Map` is a specialized map implementation that provides a concurrent, thread-safe map. It is specifically designed for use cases where the entry set of the map is stable over time, such as caches.

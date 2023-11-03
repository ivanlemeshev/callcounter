package callcounter

import (
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
		counter CallCounter
		calls   []int
	}{
		{
			name:    "mutex",
			counter: NewCallCounterMutex(),
			calls:   []int{0, 1, 3, 100},
		},
		{
			name:    "rwmutex",
			counter: NewCallCounterRWMutex(),
			calls:   []int{0, 1, 3, 100},
		},
		{
			name:    "map",
			counter: NewCallCounterMap(),
			calls:   []int{0, 1, 3, 100},
		},
		{
			name:    "channels",
			counter: NewCallCounterChannels(ctx),
			calls:   []int{0, 1, 3, 100},
		},
		{
			name:    "mutex channel",
			counter: NewCallCounterChannelMutex(),
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
		counter CallCounter
	}{
		{
			name:    "mutex",
			counter: NewCallCounterMutex(),
		},
		{
			name:    "rwmutex",
			counter: NewCallCounterRWMutex(),
		},
		{
			name:    "map",
			counter: NewCallCounterMap(),
		},
		{
			name:    "channels",
			counter: NewCallCounterChannels(ctx),
		},
		{
			name:    "mutex channel",
			counter: NewCallCounterChannelMutex(),
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

package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetInterval(t *testing.T) {
	duration := 110
	interval := 10
	expectedCountAtLeast := 10

	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Duration(duration)*time.Millisecond, cancel)

	counter := 0
	task := func(c *int) func() {
		return func() {
			*c += 1
		}
	}

	setInterval(ctx, &wg, task(&counter), time.Duration(interval)*time.Millisecond)
	wg.Wait()

	assert.LessOrEqual(t, expectedCountAtLeast, counter,
		"Expected number of task function calls with interval: %d and duration: %d should be at least: %d",
		interval, duration, expectedCountAtLeast)
}

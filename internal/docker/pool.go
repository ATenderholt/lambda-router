package docker

import (
	"context"
	"errors"
	"time"
)

type IntPool struct {
	available chan int
}

func NewIntPool(min int, max int) IntPool {
	available := make(chan int, max-min+1)

	for i := min; i <= max; i++ {
		available <- i
	}

	return IntPool{
		available: available,
	}
}

func (pool IntPool) Get(ctx context.Context) (int, error) {
	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	select {
	case x := <-pool.available:
		return x, nil
	case <-timeout.Done():
		logger.Error("Timeout while waiting for available port")
		return 0, errors.New("timeout while waiting for available port")
	}
}

func (pool IntPool) Put(value int) {
	pool.available <- value
}

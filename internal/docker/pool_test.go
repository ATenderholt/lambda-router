package docker_test

import (
	"context"
	"github.com/ATenderholt/rainbow-functions/internal/docker"
	"testing"
)

func TestPoolGet(t *testing.T) {
	pool := docker.NewIntPool(10, 15)
	for i := 0; i < 2; i++ {
		value, err := pool.Get(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if value != 10+i {
			t.Errorf("Expected %d, but got %d", 10+i, value)
		}
	}
}

func TestPoolGetAndReturn(t *testing.T) {
	// contains two values
	pool := docker.NewIntPool(10, 11)

	// get the first value
	v1, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if v1 != 10 {
		t.Fatalf("Expected 10, but got %d", v1)
	}
	pool.Put(v1)

	// get the second value
	v2, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if v2 != 11 {
		t.Fatalf("Expected 11, but got %d", v2)
	}
	pool.Put(v2)

	// next get should be the first one
	v3, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if v1 != v3 {
		t.Errorf("Expected %d, but got %d", v1, v3)
	}
}

func TestPoolExhausted(t *testing.T) {
	// contains just one values
	pool := docker.NewIntPool(10, 10)

	// get the first value
	v1, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if v1 != 10 {
		t.Fatalf("Expected 10, but got %d", v1)
	}

	v2, err := pool.Get(context.Background())
	if err == nil {
		t.Fatalf("Expected a timeout, but got a value: %d", v2)
	}
}

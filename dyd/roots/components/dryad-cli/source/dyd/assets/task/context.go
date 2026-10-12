package task

import (
	"context"
	"errors"
	"sync"
)

type contextKey uint8

const (
	semaphoreKey contextKey = iota
	cacheRegistryKey
)

var (
	ErrMissingSemaphore     = errors.New("task: context has no semaphore; configure it with task.WithConcurrency or task.NewContext")
	ErrMissingCacheRegistry = errors.New("task: context has no cache registry; configure it with task.WithCache or task.NewContext")
)

var DEFAULT_CONTEXT = NewContext(context.Background(), 16)
var SERIAL_CONTEXT = NewContext(context.Background(), 1)

// NewContext attaches a semaphore and an empty memoizer cache registry to parent.
func NewContext(parent context.Context, parallel int) context.Context {
	return WithCache(WithConcurrency(parent, parallel))
}

// WithCache derives a context with a fresh memoizer cache registry. It inherits
// the parent's semaphore, values, deadline, and cancellation. Memoizers share
// their individual caches within this registry; the parent's caches are unchanged.
func WithCache(parent context.Context) context.Context {
	return context.WithValue(parent, cacheRegistryKey, new(sync.Map))
}

// WithConcurrency derives a context with a fresh semaphore and inherits the
// parent's cache registry, values, deadline, and cancellation. The calling
// goroutine provides one execution slot; the channel bounds additional goroutines.
func WithConcurrency(parent context.Context, concurrency int) context.Context {
	if concurrency < 1 {
		panic("task: concurrency must be at least 1")
	}
	return context.WithValue(parent, semaphoreKey, make(chan struct{}, concurrency-1))
}

func semaphore(ctx context.Context) (chan struct{}, error) {
	if ctx != nil {
		if channel, ok := ctx.Value(semaphoreKey).(chan struct{}); ok && channel != nil {
			return channel, nil
		}
	}
	return nil, ErrMissingSemaphore
}

// Concurrency returns the configured execution limit, including the calling goroutine.
func Concurrency(ctx context.Context) (int, error) {
	channel, err := semaphore(ctx)
	if err != nil {
		return 0, err
	}
	return cap(channel) + 1, nil
}

func cacheRegistry(ctx context.Context) (*sync.Map, error) {
	if ctx != nil {
		if registry, ok := ctx.Value(cacheRegistryKey).(*sync.Map); ok && registry != nil {
			return registry, nil
		}
	}
	return nil, ErrMissingCacheRegistry
}

func WithContext[A any, B any](
	task Task[A, B],
	getContext Task[A, context.Context],
) Task[A, B] {
	return func(ctx context.Context, a A) (error, B) {
		err, next := getContext(ctx, a)
		if err != nil {
			return err, empty[B]()
		}
		return task(next, a)
	}
}

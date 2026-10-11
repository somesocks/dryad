package task

import (
	"context"
	"sync"
)

// A non-zero-sized identity keeps distinct memoizer instances distinct, even
// when their argument/result types and request keys are identical.
type memoizerIdentity struct {
	marker byte
}

func onceWrapper[A any, B any](task Task[A, B]) Task[A, B] {
	var once sync.Once
	var err error
	var res B
	return func(ctx context.Context, req A) (error, B) {
		once.Do(func() { err, res = task(ctx, req) })
		return err, res
	}
}

func Memoize[A any, B any](task Task[A, B], keyFunc Task[A, any]) Task[A, B] {
	memoized, _ := MemoizeWithInvalidation(task, keyFunc)
	return memoized
}

// MemoizeWithInvalidation returns a memoized task and an instance-bound task
// that deletes the cached execution for a request. Each memoizer owns a separate
// cache in the execution context's registry, allocated lazily on first use.
// Cached executions are retained until invalidation or release of the registry.
func MemoizeWithInvalidation[A any, B any](
	task Task[A, B],
	keyFunc Task[A, any],
) (Task[A, B], Task[A, any]) {
	identity := new(memoizerIdentity)

	getCache := func(ctx context.Context) (*sync.Map, error) {
		registry, err := cacheRegistry(ctx)
		if err != nil {
			return nil, err
		}
		if cache, ok := registry.Load(identity); ok {
			return cache.(*sync.Map), nil
		}
		cache, _ := registry.LoadOrStore(identity, new(sync.Map))
		return cache.(*sync.Map), nil
	}

	memoized := func(ctx context.Context, req A) (error, B) {
		cache, err := getCache(ctx)
		if err != nil {
			return err, empty[B]()
		}
		err, key := keyFunc(ctx, req)
		if err != nil {
			return err, empty[B]()
		}
		cached, ok := cache.Load(key)
		if !ok {
			cached, _ = cache.LoadOrStore(key, onceWrapper(task))
		}
		return cached.(Task[A, B])(ctx, req)
	}

	invalidate := func(ctx context.Context, req A) (error, any) {
		cache, err := getCache(ctx)
		if err != nil {
			return err, nil
		}
		err, key := keyFunc(ctx, req)
		if err != nil {
			return err, nil
		}
		cache.Delete(key)
		return nil, nil
	}

	return memoized, invalidate
}

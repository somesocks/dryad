package task_test

import (
	"context"
	"dryad/task"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestContextPreservesParent(t *testing.T) {
	type valueKey struct{}
	deadline := time.Now().Add(time.Hour)
	parent, cancel := context.WithDeadline(context.WithValue(context.Background(), valueKey{}, "caller"), deadline)
	defer cancel()
	ctx := task.NewContext(parent, 3)
	cancel()

	check := func(ctx context.Context, req int) (error, int) {
		if ctx.Value(valueKey{}) != "caller" {
			return errors.New("caller value lost"), 0
		}
		if got, ok := ctx.Deadline(); !ok || !got.Equal(deadline) {
			return errors.New("deadline lost"), 0
		}
		if !errors.Is(ctx.Err(), context.Canceled) {
			return errors.New("cancellation lost"), 0
		}
		return nil, req
	}
	wrapped := task.WithContext(
		task.ParallelMap(check),
		func(ctx context.Context, _ []int) (error, context.Context) {
			return nil, task.WithCache(task.WithConcurrency(ctx, 2))
		},
	)
	// Migration preserves cancellation information without adding cancellation
	// checks to task execution itself.
	if err, result := wrapped(ctx, []int{1, 2}); err != nil || len(result) != 2 || result[1] != 2 {
		t.Fatalf("wrapped execution: %v, %v", err, result)
	}
}

func TestMissingContextConfiguration(t *testing.T) {
	var calls int
	work := func(context.Context, int) (error, int) { calls++; return nil, 1 }
	key := func(context.Context, int) (error, any) { return nil, "key" }
	for _, ctx := range []context.Context{context.Background(), nil} {
		if err, _ := task.Parallel(work, work)(ctx, 0); !errors.Is(err, task.ErrMissingSemaphore) {
			t.Fatalf("parallel error: %v", err)
		}
		if err, _ := task.ParallelMap(work)(ctx, []int{1}); !errors.Is(err, task.ErrMissingSemaphore) {
			t.Fatalf("parallel map error: %v", err)
		}
		if err, _ := task.Memoize(work, key)(ctx, 0); !errors.Is(err, task.ErrMissingCacheRegistry) {
			t.Fatalf("memoization error: %v", err)
		}
	}
	if calls != 0 {
		t.Fatalf("unconfigured tasks executed %d times", calls)
	}
}

func TestReturnSuppliesConfiguredContext(t *testing.T) {
	work := task.Memoize(
		func(context.Context, int) (error, int) { return nil, 42 },
		func(context.Context, int) (error, any) { return nil, "key" },
	)
	run := task.Return(task.Parallel1(work), func(err error, result task.Tuple1[int]) int {
		if err != nil {
			t.Errorf("returned task failed: %v", err)
		}
		return result.A
	})
	if got := run(0); got != 42 {
		t.Fatalf("result: %d", got)
	}
}

func TestCacheLayerReuseAndIsolation(t *testing.T) {
	parent := task.NewContext(context.Background(), 3)
	var calls atomic.Int32
	work := task.Memoize(
		func(context.Context, int) (error, int32) { return nil, calls.Add(1) },
		func(_ context.Context, req int) (error, any) { return nil, req },
	)
	check := func(ctx context.Context, want int32) {
		t.Helper()
		for range 2 {
			if err, value := work(ctx, 0); err != nil || value != want {
				t.Fatalf("execution: %v, %d; want %d", err, value, want)
			}
		}
	}
	check(parent, 1)
	child := task.WithCache(parent)
	check(child, 2)
	check(context.WithValue(child, struct{}{}, "nested"), 2)
	check(task.WithConcurrency(parent, 1), 1)
	check(task.WithCache(parent), 3)
	check(parent, 1)

	wrapped := task.WithContext(task.Parallel(work, work), func(ctx context.Context, _ int) (error, context.Context) {
		return nil, task.WithCache(ctx)
	})
	for _, want := range []int32{4, 5} {
		if err, values := wrapped(parent, 0); err != nil || values.A != want || values.B != want {
			t.Fatalf("wrapped execution: %v, %+v; want %d", err, values, want)
		}
	}
	check(parent, 1)
}

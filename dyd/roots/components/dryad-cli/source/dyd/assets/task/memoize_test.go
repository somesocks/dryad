package task_test

import (
	"context"
	"dryad/task"
	"errors"
	"sync/atomic"
	"testing"
)

func memoKey(_ context.Context, req int) (error, any) { return nil, req }

func TestMemoizersOwnSeparateCaches(t *testing.T) {
	ctx := task.NewContext(context.Background(), 4)
	var callsA, callsB int
	a := task.Memoize(func(context.Context, int) (error, int) { callsA++; return nil, 10 }, memoKey)
	b := task.Memoize(func(context.Context, int) (error, int) { callsB++; return nil, 20 }, memoKey)
	c := task.Memoize(func(context.Context, int) (error, string) { return nil, "different type" }, memoKey)
	for range 2 {
		if err, got := a(ctx, 1); err != nil || got != 10 {
			t.Fatalf("A: %v, %v", err, got)
		}
		if err, got := b(ctx, 1); err != nil || got != 20 {
			t.Fatalf("B: %v, %v", err, got)
		}
		if err, got := c(ctx, 1); err != nil || got != "different type" {
			t.Fatalf("C: %v, %v", err, got)
		}
	}
	if callsA != 1 || callsB != 1 {
		t.Fatalf("calls: A=%d B=%d", callsA, callsB)
	}

	// A normal derived context inherits the same registry. A new execution
	// context instead gets independent caches, even for the same memoizer.
	derived := context.WithValue(ctx, struct{}{}, "value")
	a(derived, 1)
	a(task.NewContext(ctx, 1), 1)
	if callsA != 2 {
		t.Fatalf("calls after deriving contexts: %d", callsA)
	}
}

func TestMemoizeConcurrentSiblingCalls(t *testing.T) {
	ctx := task.NewContext(context.Background(), 8)
	var calls atomic.Int32
	work := task.Memoize(func(_ context.Context, req int) (error, int) {
		calls.Add(1)
		return nil, req * 2
	}, memoKey)
	requests := make([]int, 64)
	for i := range requests {
		requests[i] = i % 4
	}
	err, results := task.ParallelMap(work)(ctx, requests)
	if err != nil {
		t.Fatal(err)
	}
	for i, got := range results {
		if got != requests[i]*2 {
			t.Fatalf("result[%d]: %d", i, got)
		}
	}
	if got := calls.Load(); got != 4 {
		t.Fatalf("executions: %d, want 4", got)
	}
}

func TestMemoizeRecursiveCalls(t *testing.T) {
	ctx := task.NewContext(context.Background(), 2)
	calls := make(map[int]int)
	var fib task.Task[int, int]
	fib = task.Memoize(func(ctx context.Context, n int) (error, int) {
		calls[n]++
		if n < 2 {
			return nil, n
		}
		err, a := fib(ctx, n-1)
		if err != nil {
			return err, 0
		}
		err, b := fib(ctx, n-2)
		return err, a + b
	}, memoKey)
	if err, got := fib(ctx, 10); err != nil || got != 55 {
		t.Fatalf("fib: %v, %d", err, got)
	}
	for n := 0; n <= 10; n++ {
		if calls[n] != 1 {
			t.Fatalf("fib(%d) executed %d times", n, calls[n])
		}
	}
}

func TestMemoizeInvalidationIsInstanceScoped(t *testing.T) {
	ctx := task.NewContext(context.Background(), 1)
	calls := make(map[int]int)
	a, invalidate := task.MemoizeWithInvalidation(func(_ context.Context, req int) (error, int) {
		calls[req]++
		return nil, calls[req]
	}, memoKey)
	var callsB int
	b := task.Memoize(func(context.Context, int) (error, int) { callsB++; return nil, callsB }, memoKey)
	a(ctx, 1)
	a(ctx, 2)
	b(ctx, 1)
	if err, _ := invalidate(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if err, got := a(ctx, 1); err != nil || got != 2 {
		t.Fatalf("invalidated A: %v, %d", err, got)
	}
	if err, got := a(ctx, 2); err != nil || got != 1 {
		t.Fatalf("untouched A: %v, %d", err, got)
	}
	if err, got := b(ctx, 1); err != nil || got != 1 {
		t.Fatalf("untouched B: %v, %d", err, got)
	}
}

func TestMemoizeCachesExecutionErrors(t *testing.T) {
	ctx := task.NewContext(context.Background(), 1)
	want := errors.New("failure")
	var calls int
	work := task.Memoize(func(context.Context, int) (error, int) { calls++; return want, 7 }, memoKey)
	for range 2 {
		if err, got := work(ctx, 0); !errors.Is(err, want) || got != 7 {
			t.Fatalf("execution: %v, %d", err, got)
		}
	}
	if calls != 1 {
		t.Fatalf("executions: %d", calls)
	}
}

func TestMemoizeKeyErrorsDoNotExecute(t *testing.T) {
	ctx := task.NewContext(context.Background(), 1)
	want := errors.New("bad key")
	var calls int
	work := task.Memoize(
		func(context.Context, int) (error, int) { calls++; return nil, 1 },
		func(context.Context, int) (error, any) { return want, nil },
	)
	if err, _ := work(ctx, 0); !errors.Is(err, want) {
		t.Fatalf("key error: %v", err)
	}
	if calls != 0 {
		t.Fatalf("executions: %d", calls)
	}
}

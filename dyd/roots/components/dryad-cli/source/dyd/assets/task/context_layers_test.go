package task

import (
	"context"
	"testing"
)

func TestContextLayersReplaceStateIndependently(t *testing.T) {
	parent := NewContext(context.Background(), 3)
	parentSemaphore, err := semaphore(parent)
	if err != nil {
		t.Fatal(err)
	}
	parentRegistry, err := cacheRegistry(parent)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name            string
		ctx             context.Context
		sharesSemaphore bool
		sharesRegistry  bool
		concurrency     int
	}{
		{"cache", WithCache(parent), true, false, 3},
		{"concurrency", WithConcurrency(parent, 5), false, true, 5},
		{"both", NewContext(parent, 2), false, false, 2},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			channel, err := semaphore(test.ctx)
			if err != nil {
				t.Fatal(err)
			}
			registry, err := cacheRegistry(test.ctx)
			if err != nil {
				t.Fatal(err)
			}
			if (channel == parentSemaphore) != test.sharesSemaphore {
				t.Fatal("unexpected semaphore sharing")
			}
			if (registry == parentRegistry) != test.sharesRegistry {
				t.Fatal("unexpected cache registry sharing")
			}
			if cap(channel)+1 != test.concurrency {
				t.Fatalf("concurrency: %d", cap(channel)+1)
			}
		})
	}
	// Replacing child state leaves the parent's state intact.
	channel, _ := semaphore(parent)
	registry, _ := cacheRegistry(parent)
	if channel != parentSemaphore || registry != parentRegistry {
		t.Fatal("parent state changed")
	}
}

func TestOperationCacheLayersDoNotRegisterInParent(t *testing.T) {
	parent := NewContext(context.Background(), 1)
	registry, err := cacheRegistry(parent)
	if err != nil {
		t.Fatal(err)
	}
	work := Memoize(
		func(context.Context, int) (error, int) { return nil, 42 },
		func(_ context.Context, req int) (error, any) { return nil, req },
	)
	wrapped := WithContext(work, func(ctx context.Context, _ int) (error, context.Context) {
		return nil, WithCache(ctx)
	})
	for i := range 20 {
		if err, value := wrapped(parent, i); err != nil || value != 42 {
			t.Fatalf("execution: %v, %d", err, value)
		}
	}
	registry.Range(func(_, _ any) bool { t.Error("operation cache registered in parent"); return false })
}

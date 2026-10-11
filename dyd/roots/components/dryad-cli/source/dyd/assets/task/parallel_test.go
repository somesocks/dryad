package task_test

import (
	"context"
	"dryad/task"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestSerialMapStopsAfterError(t *testing.T) {
	ctx := task.NewContext(context.Background(), 1)
	want := errors.New("stop")
	var visited []int
	work := func(_ context.Context, req int) (error, int) {
		visited = append(visited, req)
		if req == 2 {
			return want, req
		}
		return nil, req
	}
	if err, _ := task.ParallelMap(work)(ctx, []int{1, 2, 3}); !errors.Is(err, want) {
		t.Fatalf("error: %v", err)
	}
	if !reflect.DeepEqual(visited, []int{1, 2}) {
		t.Fatalf("execution order: %v", visited)
	}
}

func TestParallelWaitsForResultsAndReleasesSlots(t *testing.T) {
	ctx := task.NewContext(context.Background(), 3)
	started := make(chan struct{}, 3)
	release := make(chan struct{})
	work := func(_ context.Context, req int) (error, int) {
		started <- struct{}{}
		<-release
		return nil, req * 2
	}
	type outcome struct {
		err    error
		values []int
	}
	done := make(chan outcome, 1)
	go func() {
		err, values := task.ParallelMap(work)(ctx, []int{1, 2, 3})
		done <- outcome{err, values}
	}()
	for range 3 {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			close(release)
			t.Fatal("three execution slots were not available")
		}
	}
	select {
	case <-done:
		t.Fatal("parallel task returned before its children completed")
	default:
	}
	close(release)
	result := <-done
	if result.err != nil || !reflect.DeepEqual(result.values, []int{2, 4, 6}) {
		t.Fatalf("result: %v, %v", result.err, result.values)
	}
	// All semaphore slots must be reusable when the call returns.
	ready := make(chan struct{}, 3)
	gate := make(chan struct{})
	go func() {
		err, values := task.ParallelMap(func(_ context.Context, req int) (error, int) {
			ready <- struct{}{}
			<-gate
			return nil, req
		})(ctx, []int{4, 5, 6})
		done <- outcome{err, values}
	}()
	for range 3 {
		select {
		case <-ready:
		case <-time.After(5 * time.Second):
			close(gate)
			t.Fatal("semaphore slots were not released")
		}
	}
	close(gate)
	result = <-done
	if result.err != nil || !reflect.DeepEqual(result.values, []int{4, 5, 6}) {
		t.Fatalf("reuse: %+v", result)
	}
}

func TestNestedParallelTasksDoNotDeadlock(t *testing.T) {
	ctx := task.NewContext(context.Background(), 2)
	leaf := func(_ context.Context, req int) (error, int) { return nil, req + 1 }
	nested := task.ParallelMap(task.ParallelMap(leaf))
	done := make(chan error, 1)
	go func() {
		err, values := nested(ctx, [][]int{{1, 2}, {3, 4}, {5, 6}})
		if err == nil && !reflect.DeepEqual(values, [][]int{{2, 3}, {4, 5}, {6, 7}}) {
			err = errors.New("nested results incorrect")
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("nested parallel tasks deadlocked")
	}
}

func TestParallelConcurrentErrors(t *testing.T) {
	ctx := task.NewContext(context.Background(), 3)
	want := errors.New("child failure")
	started := make(chan struct{}, 3)
	release := make(chan struct{})
	work := func(context.Context, int) (error, int) {
		started <- struct{}{}
		<-release
		return want, 7
	}
	done := make(chan error, 1)
	go func() {
		err, values := task.Parallel3(work, work, work)(ctx, 0)
		if values.A != 7 || values.B != 7 || values.C != 7 {
			err = errors.New("partial results lost")
		}
		done <- err
	}()
	for range 3 {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			close(release)
			t.Fatal("parallel tasks did not start")
		}
	}
	close(release)
	if err := <-done; !errors.Is(err, want) {
		t.Fatalf("parallel error: %v", err)
	}
}

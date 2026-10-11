package task

import (
	"context"
	"sync"
)

type parallelGroup struct {
	channel chan struct{}
	wg      sync.WaitGroup
	mutex   sync.Mutex
	err     error
}

func newParallelGroup(ctx context.Context) (*parallelGroup, error) {
	channel, err := semaphore(ctx)
	if err != nil {
		return nil, err
	}
	return &parallelGroup{channel: channel}, nil
}

func (group *parallelGroup) error() error {
	group.mutex.Lock()
	defer group.mutex.Unlock()
	return group.err
}

func (group *parallelGroup) run(work func() error) {
	if group.error() != nil {
		return
	}
	execute := func() {
		if err := work(); err != nil {
			group.mutex.Lock()
			if group.err == nil {
				group.err = err
			}
			group.mutex.Unlock()
		}
	}
	// Never wait for a slot: nested parallel tasks must be able to execute
	// inline when their ancestors have already occupied the semaphore.
	select {
	case group.channel <- struct{}{}:
		group.wg.Add(1)
		go func() {
			defer group.wg.Done()
			defer func() { <-group.channel }()
			execute()
		}()
	default:
		execute()
	}
}

func (group *parallelGroup) wait() error {
	group.wg.Wait()
	return group.error()
}

func parallelRun[A any, B any](group *parallelGroup, ctx context.Context, req A, task Task[A, B], result *B) {
	group.run(func() error {
		err, res := task(ctx, req)
		*result = res
		return err
	})
}

func Parallel1[A any, B any](ab Task[A, B]) Task[A, Tuple1[B]] {
	return func(ctx context.Context, a A) (error, Tuple1[B]) {
		var res Tuple1[B]
		group, err := newParallelGroup(ctx)
		if err != nil {
			return err, res
		}
		parallelRun(group, ctx, a, ab, &res.A)
		err = group.wait()
		return err, res
	}
}

func Parallel[A any, B any, C any](ab Task[A, B], ac Task[A, C]) Task[A, Tuple2[B, C]] {
	return func(ctx context.Context, a A) (error, Tuple2[B, C]) {
		var res Tuple2[B, C]
		group, err := newParallelGroup(ctx)
		if err != nil {
			return err, res
		}
		parallelRun(group, ctx, a, ab, &res.A)
		parallelRun(group, ctx, a, ac, &res.B)
		err = group.wait()
		return err, res
	}
}

func Parallel2[A any, B any, C any](ab Task[A, B], ac Task[A, C]) Task[A, Tuple2[B, C]] {
	return Parallel(ab, ac)
}

func Parallel3[A any, B any, C any, D any](ab Task[A, B], ac Task[A, C], ad Task[A, D]) Task[A, Tuple3[B, C, D]] {
	return func(ctx context.Context, a A) (error, Tuple3[B, C, D]) {
		var res Tuple3[B, C, D]
		group, err := newParallelGroup(ctx)
		if err != nil {
			return err, res
		}
		parallelRun(group, ctx, a, ab, &res.A)
		parallelRun(group, ctx, a, ac, &res.B)
		parallelRun(group, ctx, a, ad, &res.C)
		err = group.wait()
		return err, res
	}
}

func Parallel4[A any, B any, C any, D any, E any](ab Task[A, B], ac Task[A, C], ad Task[A, D], ae Task[A, E]) Task[A, Tuple4[B, C, D, E]] {
	return func(ctx context.Context, a A) (error, Tuple4[B, C, D, E]) {
		var res Tuple4[B, C, D, E]
		group, err := newParallelGroup(ctx)
		if err != nil {
			return err, res
		}
		parallelRun(group, ctx, a, ab, &res.A)
		parallelRun(group, ctx, a, ac, &res.B)
		parallelRun(group, ctx, a, ad, &res.C)
		parallelRun(group, ctx, a, ae, &res.D)
		err = group.wait()
		return err, res
	}
}

func Parallel5[A any, B any, C any, D any, E any, F any](ab Task[A, B], ac Task[A, C], ad Task[A, D], ae Task[A, E], af Task[A, F]) Task[A, Tuple5[B, C, D, E, F]] {
	return func(ctx context.Context, a A) (error, Tuple5[B, C, D, E, F]) {
		var res Tuple5[B, C, D, E, F]
		group, err := newParallelGroup(ctx)
		if err != nil {
			return err, res
		}
		parallelRun(group, ctx, a, ab, &res.A)
		parallelRun(group, ctx, a, ac, &res.B)
		parallelRun(group, ctx, a, ad, &res.C)
		parallelRun(group, ctx, a, ae, &res.D)
		parallelRun(group, ctx, a, af, &res.E)
		err = group.wait()
		return err, res
	}
}

func Parallel6[A any, B any, C any, D any, E any, F any, G any](ab Task[A, B], ac Task[A, C], ad Task[A, D], ae Task[A, E], af Task[A, F], ag Task[A, G]) Task[A, Tuple6[B, C, D, E, F, G]] {
	return func(ctx context.Context, a A) (error, Tuple6[B, C, D, E, F, G]) {
		var res Tuple6[B, C, D, E, F, G]
		group, err := newParallelGroup(ctx)
		if err != nil {
			return err, res
		}
		parallelRun(group, ctx, a, ab, &res.A)
		parallelRun(group, ctx, a, ac, &res.B)
		parallelRun(group, ctx, a, ad, &res.C)
		parallelRun(group, ctx, a, ae, &res.D)
		parallelRun(group, ctx, a, af, &res.E)
		parallelRun(group, ctx, a, ag, &res.F)
		err = group.wait()
		return err, res
	}
}

func Parallel7[A any, B any, C any, D any, E any, F any, G any, H any](ab Task[A, B], ac Task[A, C], ad Task[A, D], ae Task[A, E], af Task[A, F], ag Task[A, G], ah Task[A, H]) Task[A, Tuple7[B, C, D, E, F, G, H]] {
	return func(ctx context.Context, a A) (error, Tuple7[B, C, D, E, F, G, H]) {
		var res Tuple7[B, C, D, E, F, G, H]
		group, err := newParallelGroup(ctx)
		if err != nil {
			return err, res
		}
		parallelRun(group, ctx, a, ab, &res.A)
		parallelRun(group, ctx, a, ac, &res.B)
		parallelRun(group, ctx, a, ad, &res.C)
		parallelRun(group, ctx, a, ae, &res.D)
		parallelRun(group, ctx, a, af, &res.E)
		parallelRun(group, ctx, a, ag, &res.F)
		parallelRun(group, ctx, a, ah, &res.G)
		err = group.wait()
		return err, res
	}
}

func Parallel8[A any, B any, C any, D any, E any, F any, G any, H any, I any](ab Task[A, B], ac Task[A, C], ad Task[A, D], ae Task[A, E], af Task[A, F], ag Task[A, G], ah Task[A, H], ai Task[A, I]) Task[A, Tuple8[B, C, D, E, F, G, H, I]] {
	return func(ctx context.Context, a A) (error, Tuple8[B, C, D, E, F, G, H, I]) {
		var res Tuple8[B, C, D, E, F, G, H, I]
		group, err := newParallelGroup(ctx)
		if err != nil {
			return err, res
		}
		parallelRun(group, ctx, a, ab, &res.A)
		parallelRun(group, ctx, a, ac, &res.B)
		parallelRun(group, ctx, a, ad, &res.C)
		parallelRun(group, ctx, a, ae, &res.D)
		parallelRun(group, ctx, a, af, &res.E)
		parallelRun(group, ctx, a, ag, &res.F)
		parallelRun(group, ctx, a, ah, &res.G)
		parallelRun(group, ctx, a, ai, &res.H)
		err = group.wait()
		return err, res
	}
}

func ParallelMap[A any, B any](ab Task[A, B]) Task[[]A, []B] {
	return func(ctx context.Context, a []A) (error, []B) {
		res := make([]B, len(a))
		group, err := newParallelGroup(ctx)
		if err != nil {
			return err, res
		}
		for i := range a {
			parallelRun(group, ctx, a[i], ab, &res[i])
		}
		err = group.wait()
		return err, res
	}
}

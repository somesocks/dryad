package task

import (
	"sync"
	"sync/atomic"
)

type cacheKey[A any] struct {
	group string
	request A
}

func onceWrapper[A any, B any](
	task Task[A, B],
) Task[A, B] {
	var done atomic.Bool
	var mutex sync.Mutex

	var err error
	var res B

	var execute = func (ctx *ExecutionContext, req A) {
		mutex.Lock()
		defer mutex.Unlock()
		if done.Load() == false {
			defer done.Store(true)
			err, res = task(ctx, req)
		}
	}

	return func (ctx *ExecutionContext, req A) (error, B) {
		if done.Load() == false {
			execute(ctx, req)	
		}
		return err, res
	}

}

func Memoize[A any, B any](
	task Task[A, B],
	group string,
) Task[A, B] {
	return func (ctx *ExecutionContext, req A) (error, B) {
		var key = cacheKey[A]{
			group: group,
			request: req,
		}

		var cachedTask any 
		var hasCachedTask bool
		cachedTask, hasCachedTask = ctx.ExecutionCache.Load(key)
		if !hasCachedTask {
			cachedTask, hasCachedTask = ctx.ExecutionCache.LoadOrStore(key, onceWrapper(task)) 
		}

		var err error
		var res B
		err, res = cachedTask.(Task[A, B])(ctx, req)
		return err, res
	}
}
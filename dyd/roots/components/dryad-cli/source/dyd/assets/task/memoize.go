package task

import (
	"sync"
	"sync/atomic"
	// zlog "github.com/rs/zerolog/log"
	// "encoding/json"
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
		// zlog.Trace().
		// 	Str("memoGroup", group).
		// 	Bool("hasCachedTask", hasCachedTask).
		// 	Str("request", string(jsonReq)).
		// 	Msg("memogroup load")
		if !hasCachedTask {
			cachedTask, hasCachedTask = ctx.ExecutionCache.LoadOrStore(key, onceWrapper(task)) 
			// zlog.Trace().
			// 	Str("memoGroup", group).
			// 	Bool("hasCachedTask", hasCachedTask).
			// 	Str("request", string(jsonReq)).
			// 	Msg("memogroup updated")
		}

		var err error
		var res B
		err, res = cachedTask.(Task[A, B])(ctx, req)

		// var jsonReq, _ = json.Marshal(req)
		// var jsonRes, _ = json.Marshal(res)

		// zlog.Info().
		// 	Str("memoGroup", group).
		// 	Bool("hasCachedTask", hasCachedTask).
		// 	Str("request", string(jsonReq)).
		// 	Str("result", string(jsonRes)).
		// 	Msg("memogroup updated")
		return err, res
	}
}
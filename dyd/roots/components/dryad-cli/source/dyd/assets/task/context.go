package task

type ExecutionContext struct {
	ConcurrencyChannel chan struct{}
}

var DEFAULT_CONTEXT = &ExecutionContext{
	ConcurrencyChannel: make(chan struct{}, 8),
}

func WithContext[A any, B any](
	task Task[A, B],
	context Task[A, *ExecutionContext],
) Task[A, B] {
	return func (ctx *ExecutionContext, a A) (error, B) {
		err, ctx2 := context(ctx, a)
		if err != nil {
			return err, empty[B]()
		}
		return task(ctx2, a);
	}
}
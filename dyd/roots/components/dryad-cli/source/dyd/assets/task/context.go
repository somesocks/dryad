package task

type ExecutionContext struct {
	ConcurrencyChannel chan struct{}
}

var DEFAULT_CONTEXT = &ExecutionContext{
	ConcurrencyChannel: make(chan struct{}, 15),
}

var SERIAL_CONTEXT = &ExecutionContext{
	ConcurrencyChannel: make(chan struct{}, 0),
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

func BuildContext (
	parallel int,
) *ExecutionContext {
	return &ExecutionContext{
		ConcurrencyChannel: make(chan struct{}, parallel - 1),
	}
}
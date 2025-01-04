package task


type Task[A any, B any] func (ctx *ExecutionContext, req A) (error, B)

func From[A any, B any](
	task func(A) (error, B),
) Task[A, B] {	
	return func (ctx *ExecutionContext, req A) (error, B) {
		err, res := task(req)
		return err, res
	}
}

func empty[A any] () A {
	var a A
	return a
}


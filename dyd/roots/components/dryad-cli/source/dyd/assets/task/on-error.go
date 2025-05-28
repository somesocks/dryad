package task

func OnFailure[A any, B any] (
	task Task[A, B],
	onFailure Task[Tuple2[A, error], any],
) Task[A, B] {
	return func (ctx *ExecutionContext, req A) (error, B) {
		err, res := task(ctx, req)
		if err != nil {
			onFailure(
				ctx,
				Tuple2[A, error]{
					A: req,
					B: err,
				},
			)
		}
		return err, res
	}
}
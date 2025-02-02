package task

func Return[A any, B any, C any] (
	ab Task[A, B],
	ret func (err error, result B) C,
) func(a A) C {
	return func (a A) C {
		err, b := ab(nil, a)
		return ret(err, b)
	}
}
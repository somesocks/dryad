package tasks

type Task[A any, B any] func (value A) (error, B)

func empty[A any] () A {
	var a A
	return a
}

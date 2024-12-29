package tasks

type Task[A any, B any] func (value A) (error, B)

func empty[A any] () A {
	var a A
	return a
}


func Series0[A any] () Task[A, A] {
	return func (a A) (error, A) {
		return nil, a
	}
}

func Series1[A any, B any] (
	ab Task[A, B],
) Task[A, B] {
	return func (a A) (error, B) {
		err, b := ab(a)
		if err != nil { return err, empty[B]() }
		return nil, b
	}
}

func Series[A any, B any, C any] (
	ab Task[A, B],
	bc Task[B, C],
) Task[A, C] {
	return func (a A) (error, C) {
		err, b := ab(a)
		if err != nil { return err, empty[C]() }
		err, c := bc(b)
		if err != nil { return err, empty[C]() }
		return nil, c
	}
}

func Series2[A any, B any, C any] (
	ab Task[A, B],
	bc Task[B, C],
) Task[A, C] {
	return func (a A) (error, C) {
		err, b := ab(a)
		if err != nil { return err, empty[C]() }
		err, c := bc(b)
		if err != nil { return err, empty[C]() }
		return nil, c
	}
}

func Series3[A any, B any, C any, D any] (
	ab Task[A, B],
	bc Task[B, C],
	cd Task[C, D],
) Task[A, D] {
	return func (a A) (error, D) {
		err, b := ab(a)
		if err != nil { return err, empty[D]() }
		err, c := bc(b)
		if err != nil { return err, empty[D]() }
		err, d := cd(c)
		if err != nil { return err, empty[D]() }
		return nil, d
	}
}

func Series4[A any, B any, C any, D any, E any] (
	ab Task[A, B],
	bc Task[B, C],
	cd Task[C, D],
	de Task[D, E],
) Task[A, E] {
	return func (a A) (error, E) {
		err, b := ab(a)
		if err != nil { return err, empty[E]() }
		err, c := bc(b)
		if err != nil { return err, empty[E]() }
		err, d := cd(c)
		if err != nil { return err, empty[E]() }
		err, e := de(d)
		if err != nil { return err, empty[E]() }
		return nil, e
	}
}

func Series5[A any, B any, C any, D any, E any, F any] (
	ab Task[A, B],
	bc Task[B, C],
	cd Task[C, D],
	de Task[D, E],
	ef Task[E, F],
) Task[A, F] {
	return func (a A) (error, F) {
		err, b := ab(a)
		if err != nil { return err, empty[F]() }
		err, c := bc(b)
		if err != nil { return err, empty[F]() }
		err, d := cd(c)
		if err != nil { return err, empty[F]() }
		err, e := de(d)
		if err != nil { return err, empty[F]() }
		err, f := ef(e)
		if err != nil { return err, empty[F]() }
		return nil, f
	}
}

func Series6[A any, B any, C any, D any, E any, F any, G any] (
	ab Task[A, B],
	bc Task[B, C],
	cd Task[C, D],
	de Task[D, E],
	ef Task[E, F],
	fg Task[F, G],
) Task[A, G] {
	return func (a A) (error, G) {
		err, b := ab(a)
		if err != nil { return err, empty[G]() }
		err, c := bc(b)
		if err != nil { return err, empty[G]() }
		err, d := cd(c)
		if err != nil { return err, empty[G]() }
		err, e := de(d)
		if err != nil { return err, empty[G]() }
		err, f := ef(e)
		if err != nil { return err, empty[G]() }
		err, g := fg(f)
		if err != nil { return err, empty[G]() }
		return nil, g
	}
}

func Series7[A any, B any, C any, D any, E any, F any, G any, H any] (
	ab Task[A, B],
	bc Task[B, C],
	cd Task[C, D],
	de Task[D, E],
	ef Task[E, F],
	fg Task[F, G],
	gh Task[G, H],
) Task[A, H] {
	return func (a A) (error, H) {
		err, b := ab(a)
		if err != nil { return err, empty[H]() }
		err, c := bc(b)
		if err != nil { return err, empty[H]() }
		err, d := cd(c)
		if err != nil { return err, empty[H]() }
		err, e := de(d)
		if err != nil { return err, empty[H]() }
		err, f := ef(e)
		if err != nil { return err, empty[H]() }
		err, g := fg(f)
		if err != nil { return err, empty[H]() }
		err, h := gh(g)
		if err != nil { return err, empty[H]() }
		return nil, h
	}
}

type Tuple1[A any] struct {
	A A
}

type Tuple2[A any, B any] struct {
	A A
	B B
}

type Tuple3[A any, B any, C any] struct {
	A A
	B B
	C C
}

func Parallel1[A any, B any] (
	ab Task[A, B],
) Task[A, Tuple1[B]] {
	return func (a A) (error, Tuple1[B]) {
		err, b := ab(a)
		if err != nil { return err, empty[Tuple1[B]]() }
		return nil, Tuple1[B]{
			A: b,
		}
	}
}

func Parallel[A any, B any, C any] (
	ab Task[A, B],
	ac Task[A, C],
) Task[A, Tuple2[B, C]] {
	return func (a A) (error, Tuple2[B, C]) {
		err, b := ab(a)
		if err != nil { return err, empty[Tuple2[B, C]]() }
		err, c := ac(a)
		if err != nil { return err, empty[Tuple2[B, C]]() }
		return nil, Tuple2[B, C]{
			A: b,
			B: c,
		}
	}
}

func Parallel2[A any, B any, C any] (
	ab Task[A, B],
	ac Task[A, C],
) Task[A, Tuple2[B, C]] {
	return func (a A) (error, Tuple2[B, C]) {
		err, b := ab(a)
		if err != nil { return err, empty[Tuple2[B, C]]() }
		err, c := ac(a)
		if err != nil { return err, empty[Tuple2[B, C]]() }
		return nil, Tuple2[B, C]{
			A: b,
			B: c,
		}
	}
}

func Return[A any, B any, C any] (
	ab Task[A, B],
	ret func (err error, result B) C,
) func(a A) C {
	return func (a A) C {
		err, b := ab(a)
		return ret(err, b)
	}
}
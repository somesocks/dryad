package task


func Series1[A any, B any] (
	ab Task[A, B],
) Task[A, B] {
	return func (ctx *ExecutionContext, a A) (error, B) {
		if ctx == nil { ctx = DEFAULT_CONTEXT }

		err, b := ab(ctx, a)
		if err != nil { return err, empty[B]() }
		return nil, b
	}
}

func Series[A any, B any, C any] (
	ab Task[A, B],
	bc Task[B, C],
) Task[A, C] {
	return func (ctx *ExecutionContext, a A) (error, C) {
		if ctx == nil { ctx = DEFAULT_CONTEXT }

		err, b := ab(ctx, a)
		if err != nil { return err, empty[C]() }
		err, c := bc(ctx, b)
		if err != nil { return err, empty[C]() }
		return nil, c
	}
}

func Series2[A any, B any, C any] (
	ab Task[A, B],
	bc Task[B, C],
) Task[A, C] {
	return func (ctx *ExecutionContext, a A) (error, C) {
		if ctx == nil { ctx = DEFAULT_CONTEXT }

		err, b := ab(ctx, a)
		if err != nil { return err, empty[C]() }
		err, c := bc(ctx, b)
		if err != nil { return err, empty[C]() }
		return nil, c
	}
}

func Series3[A any, B any, C any, D any] (
	ab Task[A, B],
	bc Task[B, C],
	cd Task[C, D],
) Task[A, D] {
	return func (ctx *ExecutionContext, a A) (error, D) {
		if ctx == nil { ctx = DEFAULT_CONTEXT }

		err, b := ab(ctx, a)
		if err != nil { return err, empty[D]() }
		err, c := bc(ctx, b)
		if err != nil { return err, empty[D]() }
		err, d := cd(ctx, c)
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
	return func (ctx *ExecutionContext, a A) (error, E) {
		if ctx == nil { ctx = DEFAULT_CONTEXT }

		err, b := ab(ctx, a)
		if err != nil { return err, empty[E]() }
		err, c := bc(ctx, b)
		if err != nil { return err, empty[E]() }
		err, d := cd(ctx, c)
		if err != nil { return err, empty[E]() }
		err, e := de(ctx, d)
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
	return func (ctx *ExecutionContext, a A) (error, F) {
		if ctx == nil { ctx = DEFAULT_CONTEXT }

		err, b := ab(ctx, a)
		if err != nil { return err, empty[F]() }
		err, c := bc(ctx, b)
		if err != nil { return err, empty[F]() }
		err, d := cd(ctx, c)
		if err != nil { return err, empty[F]() }
		err, e := de(ctx, d)
		if err != nil { return err, empty[F]() }
		err, f := ef(ctx, e)
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
	return func (ctx *ExecutionContext, a A) (error, G) {
		if ctx == nil { ctx = DEFAULT_CONTEXT }

		err, b := ab(ctx, a)
		if err != nil { return err, empty[G]() }
		err, c := bc(ctx, b)
		if err != nil { return err, empty[G]() }
		err, d := cd(ctx, c)
		if err != nil { return err, empty[G]() }
		err, e := de(ctx, d)
		if err != nil { return err, empty[G]() }
		err, f := ef(ctx, e)
		if err != nil { return err, empty[G]() }
		err, g := fg(ctx, f)
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
	return func (ctx *ExecutionContext, a A) (error, H) {
		if ctx == nil { ctx = DEFAULT_CONTEXT }

		err, b := ab(ctx, a)
		if err != nil { return err, empty[H]() }
		err, c := bc(ctx, b)
		if err != nil { return err, empty[H]() }
		err, d := cd(ctx, c)
		if err != nil { return err, empty[H]() }
		err, e := de(ctx, d)
		if err != nil { return err, empty[H]() }
		err, f := ef(ctx, e)
		if err != nil { return err, empty[H]() }
		err, g := fg(ctx, f)
		if err != nil { return err, empty[H]() }
		err, h := gh(ctx, g)
		if err != nil { return err, empty[H]() }
		return nil, h
	}
}

func Series8[A any, B any, C any, D any, E any, F any, G any, H any, I any] (
	ab Task[A, B],
	bc Task[B, C],
	cd Task[C, D],
	de Task[D, E],
	ef Task[E, F],
	fg Task[F, G],
	gh Task[G, H],
	hi Task[H, I],
) Task[A, I] {
	return func (ctx *ExecutionContext, a A) (error, I) {
		if ctx == nil { ctx = DEFAULT_CONTEXT }

		err, b := ab(ctx, a)
		if err != nil { return err, empty[I]() }
		err, c := bc(ctx, b)
		if err != nil { return err, empty[I]() }
		err, d := cd(ctx, c)
		if err != nil { return err, empty[I]() }
		err, e := de(ctx, d)
		if err != nil { return err, empty[I]() }
		err, f := ef(ctx, e)
		if err != nil { return err, empty[I]() }
		err, g := fg(ctx, f)
		if err != nil { return err, empty[I]() }
		err, h := gh(ctx, g)
		if err != nil { return err, empty[I]() }
		err, i := hi(ctx, h)
		if err != nil { return err, empty[I]() }
		return nil, i
	}
}

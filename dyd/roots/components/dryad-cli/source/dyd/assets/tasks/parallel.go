package tasks


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

func Parallel3[A any, B any, C any, D any] (
	ab Task[A, B],
	ac Task[A, C],
	ad Task[A, D],
) Task[A, Tuple3[B, C, D]] {
	return func (a A) (error, Tuple3[B, C, D]) {
		err, b := ab(a)
		if err != nil { return err, empty[Tuple3[B, C, D]]() }
		err, c := ac(a)
		if err != nil { return err, empty[Tuple3[B, C, D]]() }
		err, d := ad(a)
		if err != nil { return err, empty[Tuple3[B, C, D]]() }
		return nil, Tuple3[B, C, D]{
			A: b,
			B: c,
			C: d,
		}
	}
}

func Parallel4[A any, B any, C any, D any, E any] (
	ab Task[A, B],
	ac Task[A, C],
	ad Task[A, D],
	ae Task[A, E],
) Task[A, Tuple4[B, C, D, E]] {
	return func (a A) (error, Tuple4[B, C, D, E]) {
		err, b := ab(a)
		if err != nil { return err, empty[Tuple4[B, C, D, E]]() }
		err, c := ac(a)
		if err != nil { return err, empty[Tuple4[B, C, D, E]]() }
		err, d := ad(a)
		if err != nil { return err, empty[Tuple4[B, C, D, E]]() }
		err, e := ae(a)
		if err != nil { return err, empty[Tuple4[B, C, D, E]]() }
		return nil, Tuple4[B, C, D, E]{
			A: b,
			B: c,
			C: d,
			D: e,
		}
	}
}

func Parallel5[A any, B any, C any, D any, E any, F any] (
	ab Task[A, B],
	ac Task[A, C],
	ad Task[A, D],
	ae Task[A, E],
	af Task[A, F],
) Task[A, Tuple5[B, C, D, E, F]] {
	return func (a A) (error, Tuple5[B, C, D, E, F]) {
		err, b := ab(a)
		if err != nil { return err, empty[Tuple5[B, C, D, E, F]]() }
		err, c := ac(a)
		if err != nil { return err, empty[Tuple5[B, C, D, E, F]]() }
		err, d := ad(a)
		if err != nil { return err, empty[Tuple5[B, C, D, E, F]]() }
		err, e := ae(a)
		if err != nil { return err, empty[Tuple5[B, C, D, E, F]]() }
		err, f := af(a)
		if err != nil { return err, empty[Tuple5[B, C, D, E, F]]() }
		return nil, Tuple5[B, C, D, E, F]{
			A: b,
			B: c,
			C: d,
			D: e,
			E: f,
		}
	}
}

func Parallel6[A any, B any, C any, D any, E any, F any, G any] (
	ab Task[A, B],
	ac Task[A, C],
	ad Task[A, D],
	ae Task[A, E],
	af Task[A, F],
	ag Task[A, G],
) Task[A, Tuple6[B, C, D, E, F, G]] {
	return func (a A) (error, Tuple6[B, C, D, E, F, G]) {
		err, b := ab(a)
		if err != nil { return err, empty[Tuple6[B, C, D, E, F, G]]() }
		err, c := ac(a)
		if err != nil { return err, empty[Tuple6[B, C, D, E, F, G]]() }
		err, d := ad(a)
		if err != nil { return err, empty[Tuple6[B, C, D, E, F, G]]() }
		err, e := ae(a)
		if err != nil { return err, empty[Tuple6[B, C, D, E, F, G]]() }
		err, f := af(a)
		if err != nil { return err, empty[Tuple6[B, C, D, E, F, G]]() }
		err, g := ag(a)
		if err != nil { return err, empty[Tuple6[B, C, D, E, F, G]]() }
		return nil, Tuple6[B, C, D, E, F, G]{
			A: b,
			B: c,
			C: d,
			D: e,
			E: f,
			F: g,
		}
	}
}

func Parallel7[A any, B any, C any, D any, E any, F any, G any, H any] (
	ab Task[A, B],
	ac Task[A, C],
	ad Task[A, D],
	ae Task[A, E],
	af Task[A, F],
	ag Task[A, G],
	ah Task[A, H],
) Task[A, Tuple7[B, C, D, E, F, G, H]] {
	return func (a A) (error, Tuple7[B, C, D, E, F, G, H]) {
		err, b := ab(a)
		if err != nil { return err, empty[Tuple7[B, C, D, E, F, G, H]]() }
		err, c := ac(a)
		if err != nil { return err, empty[Tuple7[B, C, D, E, F, G, H]]() }
		err, d := ad(a)
		if err != nil { return err, empty[Tuple7[B, C, D, E, F, G, H]]() }
		err, e := ae(a)
		if err != nil { return err, empty[Tuple7[B, C, D, E, F, G, H]]() }
		err, f := af(a)
		if err != nil { return err, empty[Tuple7[B, C, D, E, F, G, H]]() }
		err, g := ag(a)
		if err != nil { return err, empty[Tuple7[B, C, D, E, F, G, H]]() }
		err, h := ah(a)
		if err != nil { return err, empty[Tuple7[B, C, D, E, F, G, H]]() }
		return nil, Tuple7[B, C, D, E, F, G, H]{
			A: b,
			B: c,
			C: d,
			D: e,
			E: f,
			F: g,
			G: h,
		}
	}
}

func Parallel8[A any, B any, C any, D any, E any, F any, G any, H any, I any] (
	ab Task[A, B],
	ac Task[A, C],
	ad Task[A, D],
	ae Task[A, E],
	af Task[A, F],
	ag Task[A, G],
	ah Task[A, H],
	ai Task[A, I],
) Task[A, Tuple8[B, C, D, E, F, G, H, I]] {
	return func (a A) (error, Tuple8[B, C, D, E, F, G, H, I]) {
		err, b := ab(a)
		if err != nil { return err, empty[Tuple8[B, C, D, E, F, G, H, I]]() }
		err, c := ac(a)
		if err != nil { return err, empty[Tuple8[B, C, D, E, F, G, H, I]]() }
		err, d := ad(a)
		if err != nil { return err, empty[Tuple8[B, C, D, E, F, G, H, I]]() }
		err, e := ae(a)
		if err != nil { return err, empty[Tuple8[B, C, D, E, F, G, H, I]]() }
		err, f := af(a)
		if err != nil { return err, empty[Tuple8[B, C, D, E, F, G, H, I]]() }
		err, g := ag(a)
		if err != nil { return err, empty[Tuple8[B, C, D, E, F, G, H, I]]() }
		err, h := ah(a)
		if err != nil { return err, empty[Tuple8[B, C, D, E, F, G, H, I]]() }
		err, i := ai(a)
		if err != nil { return err, empty[Tuple8[B, C, D, E, F, G, H, I]]() }
		return nil, Tuple8[B, C, D, E, F, G, H, I]{
			A: b,
			B: c,
			C: d,
			D: e,
			E: f,
			F: g,
			G: h,
			H: i,
		}
	}
}

func ParallelMap[A any, B any] (
	ab Task[A, B],
) Task[[]A, []B] {
	return func (a []A) (error, []B) {
		b := make([]B, len(a))
		for k, v := range a {
			err, bb := ab(v)
			if err != nil { return err, empty[[]B]() }
			b[k] = bb
		}
		return nil, b
	}
}
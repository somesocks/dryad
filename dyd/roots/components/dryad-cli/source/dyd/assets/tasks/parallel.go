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

type parallelRunnerResult struct {
	ind int
	err error
	res any
}

func parallelRunner[A any, B any](
	ch chan parallelRunnerResult,
	task Task[A, B],
	req A,
	ind int,
) {
	err, res := task(req)
	ch <- parallelRunnerResult{ind: ind, err: err , res: res}
}

func Parallel[A any, B any, C any] (
	ab Task[A, B],
	ac Task[A, C],
) Task[A, Tuple2[B, C]] {
	return func (a A) (error, Tuple2[B, C]) {
		var err error
		var results Tuple2[B, C] 
		var ch chan parallelRunnerResult = make(chan parallelRunnerResult)

		go parallelRunner(ch, ab, a, 1)
		go parallelRunner(ch, ac, a, 2)

		for i := 0; i<2; i++ {
			task := <-ch
			if task.err != nil && err == nil { err = task.err }
			switch task.ind {
				case 1: results.A = task.res.(B)
				case 2: results.B = task.res.(C)
			}
		}

		return err, results
	}
}

func Parallel2[A any, B any, C any] (
	ab Task[A, B],
	ac Task[A, C],
) Task[A, Tuple2[B, C]] {
	return func (a A) (error, Tuple2[B, C]) {
		var err error
		var results Tuple2[B, C] 
		var ch chan parallelRunnerResult = make(chan parallelRunnerResult)

		go parallelRunner(ch, ab, a, 1)
		go parallelRunner(ch, ac, a, 2)

		for i := 0; i<2; i++ {
			task := <-ch
			if task.err != nil && err == nil { err = task.err }
			switch task.ind {
				case 1: results.A = task.res.(B)
				case 2: results.B = task.res.(C)
			}
		}

		return err, results
	}
}

func Parallel3[A any, B any, C any, D any] (
	ab Task[A, B],
	ac Task[A, C],
	ad Task[A, D],
) Task[A, Tuple3[B, C, D]] {
	return func (a A) (error, Tuple3[B, C, D]) {
		var err error
		var results Tuple3[B, C, D]
		var ch chan parallelRunnerResult = make(chan parallelRunnerResult)

		go parallelRunner(ch, ab, a, 1)
		go parallelRunner(ch, ac, a, 2)
		go parallelRunner(ch, ad, a, 3)

		for i := 0; i < 3; i++ {
			task := <-ch
			if task.err != nil && err == nil { err = task.err }
			switch task.ind {
				case 1: results.A = task.res.(B)
				case 2: results.B = task.res.(C)
				case 3: results.C = task.res.(D)
			}
		}

		return err, results
	}
}

func Parallel4[A any, B any, C any, D any, E any] (
	ab Task[A, B],
	ac Task[A, C],
	ad Task[A, D],
	ae Task[A, E],
) Task[A, Tuple4[B, C, D, E]] {
	return func (a A) (error, Tuple4[B, C, D, E]) {
		var err error
		var results Tuple4[B, C, D, E]
		var ch chan parallelRunnerResult = make(chan parallelRunnerResult)

		go parallelRunner(ch, ab, a, 1)
		go parallelRunner(ch, ac, a, 2)
		go parallelRunner(ch, ad, a, 3)
		go parallelRunner(ch, ae, a, 4)

		for i := 0; i < 4; i++ {
			task := <-ch
			if task.err != nil && err == nil { err = task.err }
			switch task.ind {
				case 1: results.A = task.res.(B)
				case 2: results.B = task.res.(C)
				case 3: results.C = task.res.(D)
				case 4: results.D = task.res.(E)
			}
		}

		return err, results
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
		var err error
		var results Tuple5[B, C, D, E, F]
		var ch chan parallelRunnerResult = make(chan parallelRunnerResult)

		go parallelRunner(ch, ab, a, 1)
		go parallelRunner(ch, ac, a, 2)
		go parallelRunner(ch, ad, a, 3)
		go parallelRunner(ch, ae, a, 4)
		go parallelRunner(ch, af, a, 5)

		for i := 0; i < 5; i++ {
			task := <-ch
			if task.err != nil && err == nil { err = task.err }
			switch task.ind {
				case 1: results.A = task.res.(B)
				case 2: results.B = task.res.(C)
				case 3: results.C = task.res.(D)
				case 4: results.D = task.res.(E)
				case 5: results.E = task.res.(F)
			}
		}

		return err, results
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
		var err error
		var results Tuple6[B, C, D, E, F, G]
		var ch chan parallelRunnerResult = make(chan parallelRunnerResult)

		go parallelRunner(ch, ab, a, 1)
		go parallelRunner(ch, ac, a, 2)
		go parallelRunner(ch, ad, a, 3)
		go parallelRunner(ch, ae, a, 4)
		go parallelRunner(ch, af, a, 5)
		go parallelRunner(ch, ag, a, 6)

		for i := 0; i < 6; i++ {
			task := <-ch
			if task.err != nil && err == nil { err = task.err }
			switch task.ind {
				case 1: results.A = task.res.(B)
				case 2: results.B = task.res.(C)
				case 3: results.C = task.res.(D)
				case 4: results.D = task.res.(E)
				case 5: results.E = task.res.(F)
				case 6: results.F = task.res.(G)
			}
		}

		return err, results
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
		var err error
		var results Tuple7[B, C, D, E, F, G, H]
		var ch chan parallelRunnerResult = make(chan parallelRunnerResult)

		go parallelRunner(ch, ab, a, 1)
		go parallelRunner(ch, ac, a, 2)
		go parallelRunner(ch, ad, a, 3)
		go parallelRunner(ch, ae, a, 4)
		go parallelRunner(ch, af, a, 5)
		go parallelRunner(ch, ag, a, 6)
		go parallelRunner(ch, ah, a, 7)

		for i := 0; i < 7; i++ {
			task := <-ch
			if task.err != nil && err == nil { err = task.err }
			switch task.ind {
				case 1: results.A = task.res.(B)
				case 2: results.B = task.res.(C)
				case 3: results.C = task.res.(D)
				case 4: results.D = task.res.(E)
				case 5: results.E = task.res.(F)
				case 6: results.F = task.res.(G)
				case 7: results.G = task.res.(H)
			}
		}

		return err, results
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
		var err error
		var results Tuple8[B, C, D, E, F, G, H, I]
		var ch chan parallelRunnerResult = make(chan parallelRunnerResult)

		go parallelRunner(ch, ab, a, 1)
		go parallelRunner(ch, ac, a, 2)
		go parallelRunner(ch, ad, a, 3)
		go parallelRunner(ch, ae, a, 4)
		go parallelRunner(ch, af, a, 5)
		go parallelRunner(ch, ag, a, 6)
		go parallelRunner(ch, ah, a, 7)
		go parallelRunner(ch, ai, a, 8)

		for i := 0; i < 8; i++ {
			task := <-ch
			if task.err != nil && err == nil { err = task.err }
			switch task.ind {
				case 1: results.A = task.res.(B)
				case 2: results.B = task.res.(C)
				case 3: results.C = task.res.(D)
				case 4: results.D = task.res.(E)
				case 5: results.E = task.res.(F)
				case 6: results.F = task.res.(G)
				case 7: results.G = task.res.(H)
				case 8: results.H = task.res.(I)
			}
		}

		return err, results
	}
}

func ParallelMap[A any, B any] (
	ab Task[A, B],
) Task[[]A, []B] {
	return func (a []A) (error, []B) {
		var b []B = make([]B, len(a))
		var err error
		var ch chan parallelRunnerResult = make(chan parallelRunnerResult)

		for i := 0; i < len(a); i++ {
			go parallelRunner(ch, ab, a[i], i)
		}

		for i := 0; i < len(a); i++ {
			task := <-ch
			if task.err != nil && err == nil { err = task.err }
			b[i] = task.res.(B)
		}

		return err, b
	}
}
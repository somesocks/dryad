package task

import (
	"sync"
)

func Parallel1[A any, B any] (
	ab Task[A, B],
) Task[A, Tuple1[B]] {
	return func (ctx *ExecutionContext, a A) (error, Tuple1[B]) {
		var err error
		var res Tuple1[B]
		var wg sync.WaitGroup

		if ctx == nil { ctx = DEFAULT_CONTEXT }

		select {
		case ctx.ConcurrencyChannel <- struct{}{}:
			wg.Add(1)
			go func () {
				err2, res2 := ab(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.A = res2
				wg.Done()
				<- ctx.ConcurrencyChannel
			}()
		default:
			err2, res2 := ab(ctx, a)
			if err2 != nil && err != nil { err = err2 }
			res.A = res2
		}
		
		wg.Wait()

		return err, res
	}
}

func Parallel[A any, B any, C any] (
	ab Task[A, B],
	ac Task[A, C],
) Task[A, Tuple2[B, C]] {
	return func (ctx *ExecutionContext, a A) (error, Tuple2[B, C]) {
		var err error
		var res Tuple2[B, C]
		var wg sync.WaitGroup

		if ctx == nil { ctx = DEFAULT_CONTEXT }

		select {
		case ctx.ConcurrencyChannel <- struct{}{}:
			wg.Add(1)
			go func () {
				err2, res2 := ab(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.A = res2
				wg.Done()
				<- ctx.ConcurrencyChannel
			}()
		default:
			err2, res2 := ab(ctx, a)
			if err2 != nil && err != nil { err = err2 }
			res.A = res2
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ac(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.B = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ac(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.B = res2
			}
		}

		wg.Wait()

		return err, res
	}
}

func Parallel2[A any, B any, C any] (
	ab Task[A, B],
	ac Task[A, C],
) Task[A, Tuple2[B, C]] {
	return func (ctx *ExecutionContext, a A) (error, Tuple2[B, C]) {
		var err error
		var res Tuple2[B, C]
		var wg sync.WaitGroup

		if ctx == nil { ctx = DEFAULT_CONTEXT }

		select {
		case ctx.ConcurrencyChannel <- struct{}{}:
			wg.Add(1)
			go func () {
				err2, res2 := ab(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.A = res2
				wg.Done()
				<- ctx.ConcurrencyChannel
			}()
		default:
			err2, res2 := ab(ctx, a)
			if err2 != nil && err != nil { err = err2 }
			res.A = res2
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ac(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.B = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ac(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.B = res2
			}
		}

		wg.Wait()

		return err, res
	}
}

func Parallel3[A any, B any, C any, D any] (
	ab Task[A, B],
	ac Task[A, C],
	ad Task[A, D],
) Task[A, Tuple3[B, C, D]] {
	return func (ctx *ExecutionContext, a A) (error, Tuple3[B, C, D]) {
		var err error
		var res Tuple3[B, C, D]
		var wg sync.WaitGroup

		if ctx == nil { ctx = DEFAULT_CONTEXT }

		select {
		case ctx.ConcurrencyChannel <- struct{}{}:
			wg.Add(1)
			go func () {
				err2, res2 := ab(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.A = res2
				wg.Done()
				<- ctx.ConcurrencyChannel
			}()
		default:
			err2, res2 := ab(ctx, a)
			if err2 != nil && err != nil { err = err2 }
			res.A = res2
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ac(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.B = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ac(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.B = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ad(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.C = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ad(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.C = res2
			}
		}

		wg.Wait()

		return err, res
	}
}

func Parallel4[A any, B any, C any, D any, E any] (
	ab Task[A, B],
	ac Task[A, C],
	ad Task[A, D],
	ae Task[A, E],
) Task[A, Tuple4[B, C, D, E]] {
	return func (ctx *ExecutionContext, a A) (error, Tuple4[B, C, D, E]) {
		var err error
		var res Tuple4[B, C, D, E]
		var wg sync.WaitGroup

		if ctx == nil { ctx = DEFAULT_CONTEXT }

		select {
		case ctx.ConcurrencyChannel <- struct{}{}:
			wg.Add(1)
			go func () {
				err2, res2 := ab(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.A = res2
				wg.Done()
				<- ctx.ConcurrencyChannel
			}()
		default:
			err2, res2 := ab(ctx, a)
			if err2 != nil && err != nil { err = err2 }
			res.A = res2
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ac(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.B = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ac(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.B = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ad(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.C = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ad(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.C = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ae(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.D = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ae(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.D = res2
			}
		}

		wg.Wait()

		return err, res
	}
}

func Parallel5[A any, B any, C any, D any, E any, F any] (
	ab Task[A, B],
	ac Task[A, C],
	ad Task[A, D],
	ae Task[A, E],
	af Task[A, F],
) Task[A, Tuple5[B, C, D, E, F]] {
	return func (ctx *ExecutionContext, a A) (error, Tuple5[B, C, D, E, F]) {
		var err error
		var res Tuple5[B, C, D, E, F]
		var wg sync.WaitGroup

		if ctx == nil { ctx = DEFAULT_CONTEXT }

		select {
		case ctx.ConcurrencyChannel <- struct{}{}:
			wg.Add(1)
			go func () {
				err2, res2 := ab(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.A = res2
				wg.Done()
				<- ctx.ConcurrencyChannel
			}()
		default:
			err2, res2 := ab(ctx, a)
			if err2 != nil && err != nil { err = err2 }
			res.A = res2
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ac(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.B = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ac(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.B = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ad(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.C = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ad(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.C = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ae(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.D = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ae(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.D = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := af(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.E = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := af(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.E = res2
			}
		}

		wg.Wait()

		return err, res
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
	return func (ctx *ExecutionContext, a A) (error, Tuple6[B, C, D, E, F, G]) {
		var err error
		var res Tuple6[B, C, D, E, F, G]
		var wg sync.WaitGroup

		if ctx == nil { ctx = DEFAULT_CONTEXT }

		select {
		case ctx.ConcurrencyChannel <- struct{}{}:
			wg.Add(1)
			go func () {
				err2, res2 := ab(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.A = res2
				wg.Done()
				<- ctx.ConcurrencyChannel
			}()
		default:
			err2, res2 := ab(ctx, a)
			if err2 != nil && err != nil { err = err2 }
			res.A = res2
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ac(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.B = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ac(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.B = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ad(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.C = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ad(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.C = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ae(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.D = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ae(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.D = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := af(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.E = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := af(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.E = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ag(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.F = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ag(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.F = res2
			}
		}

		wg.Wait()

		return err, res
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
	return func (ctx *ExecutionContext, a A) (error, Tuple7[B, C, D, E, F, G, H]) {
		var err error
		var res Tuple7[B, C, D, E, F, G, H]
		var wg sync.WaitGroup

		if ctx == nil { ctx = DEFAULT_CONTEXT }

		select {
		case ctx.ConcurrencyChannel <- struct{}{}:
			wg.Add(1)
			go func () {
				err2, res2 := ab(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.A = res2
				wg.Done()
				<- ctx.ConcurrencyChannel
			}()
		default:
			err2, res2 := ab(ctx, a)
			if err2 != nil && err != nil { err = err2 }
			res.A = res2
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ac(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.B = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ac(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.B = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ad(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.C = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ad(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.C = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ae(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.D = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ae(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.D = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := af(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.E = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := af(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.E = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ag(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.F = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ag(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.F = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ah(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.G = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ah(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.G = res2
			}
		}

		wg.Wait()

		return err, res
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
	return func (ctx *ExecutionContext, a A) (error, Tuple8[B, C, D, E, F, G, H, I]) {
		var err error
		var res Tuple8[B, C, D, E, F, G, H, I]
		var wg sync.WaitGroup

		if ctx == nil { ctx = DEFAULT_CONTEXT }

		select {
		case ctx.ConcurrencyChannel <- struct{}{}:
			wg.Add(1)
			go func () {
				err2, res2 := ab(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.A = res2
				wg.Done()
				<- ctx.ConcurrencyChannel
			}()
		default:
			err2, res2 := ab(ctx, a)
			if err2 != nil && err != nil { err = err2 }
			res.A = res2
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ac(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.B = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ac(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.B = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ad(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.C = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ad(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.C = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ae(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.D = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ae(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.D = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := af(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.E = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := af(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.E = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ag(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.F = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ag(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.F = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ah(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.G = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ah(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.G = res2
			}
		}

		if err == nil {
			select {
			case ctx.ConcurrencyChannel <- struct{}{}:
				wg.Add(1)
				go func () {
					err2, res2 := ai(ctx, a)
					if err2 != nil && err != nil { err = err2 }
					res.H = res2
					wg.Done()
					<- ctx.ConcurrencyChannel
				}()
			default:
				err2, res2 := ai(ctx, a)
				if err2 != nil && err != nil { err = err2 }
				res.H = res2
			}
		}

		wg.Wait()

		return err, res
	}
}

func ParallelMap[A any, B any] (
	ab Task[A, B],
) Task[[]A, []B] {
	return func (ctx *ExecutionContext, a []A) (error, []B) {
		var res []B = make([]B, len(a))
		var err error
		var wg sync.WaitGroup

		if ctx == nil { ctx = DEFAULT_CONTEXT }

		for i := 0; i < len(a); i++ {
			if err == nil {
				select {
				case ctx.ConcurrencyChannel <- struct{}{}:
					wg.Add(1)
					go func (ii int) {
						err2, res2 := ab(ctx, a[ii])
						if err2 != nil && err != nil { err = err2 }
						res[ii] = res2
						wg.Done()
						<- ctx.ConcurrencyChannel
					}(i)
				default:
					err2, res2 := ab(ctx, a[i])
					if err2 != nil && err != nil { err = err2 }
					res[i] = res2
				}
			}	
		}

		wg.Wait()

		return err, res
	}
}
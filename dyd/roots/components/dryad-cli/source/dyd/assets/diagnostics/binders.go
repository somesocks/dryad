package diagnostics

import "sync/atomic"

type CallA0R0 struct {
	Key string
}

type NextA0R0 func(CallA0R0) error
type RuleA0R0 func(next NextA0R0) NextA0R0
type DecoratorA0R0 = RuleA0R0

type CallA1R0[A0 any] struct {
	Key string
	A0  A0
}

type NextA1R0[A0 any] func(CallA1R0[A0]) error
type RuleA1R0[A0 any] func(next NextA1R0[A0]) NextA1R0[A0]
type DecoratorA1R0[A0 any] = RuleA1R0[A0]

type CallA2R0[A0, A1 any] struct {
	Key string
	A0  A0
	A1  A1
}

type NextA2R0[A0, A1 any] func(CallA2R0[A0, A1]) error
type RuleA2R0[A0, A1 any] func(next NextA2R0[A0, A1]) NextA2R0[A0, A1]
type DecoratorA2R0[A0, A1 any] = RuleA2R0[A0, A1]

type CallA0R1 struct {
	Key string
}

type NextA0R1[R0 any] func(CallA0R1) (error, R0)
type RuleA0R1[R0 any] func(next NextA0R1[R0]) NextA0R1[R0]
type DecoratorA0R1[R0 any] = RuleA0R1[R0]

type CallA1R1[A0 any] struct {
	Key string
	A0  A0
}

type NextA1R1[A0, R0 any] func(CallA1R1[A0]) (error, R0)
type RuleA1R1[A0, R0 any] func(next NextA1R1[A0, R0]) NextA1R1[A0, R0]
type DecoratorA1R1[A0, R0 any] = RuleA1R1[A0, R0]

type CallA2R1[A0, A1 any] struct {
	Key string
	A0  A0
	A1  A1
}

type NextA2R1[A0, A1, R0 any] func(CallA2R1[A0, A1]) (error, R0)
type RuleA2R1[A0, A1, R0 any] func(next NextA2R1[A0, A1, R0]) NextA2R1[A0, A1, R0]
type DecoratorA2R1[A0, A1, R0 any] = RuleA2R1[A0, A1, R0]

func runnerDecoratorA0R0(r runner) DecoratorA0R0 {
	return func(next NextA0R0) NextA0R0 {
		return func(call CallA0R0) error {
			if err := r(call.Key); err != nil {
				return err
			}
			return next(call)
		}
	}
}

func runnerDecoratorA1R0[A0 any](r runner) DecoratorA1R0[A0] {
	return func(next NextA1R0[A0]) NextA1R0[A0] {
		return func(call CallA1R0[A0]) error {
			if err := r(call.Key); err != nil {
				return err
			}
			return next(call)
		}
	}
}

func runnerDecoratorA2R0[A0, A1 any](r runner) DecoratorA2R0[A0, A1] {
	return func(next NextA2R0[A0, A1]) NextA2R0[A0, A1] {
		return func(call CallA2R0[A0, A1]) error {
			if err := r(call.Key); err != nil {
				return err
			}
			return next(call)
		}
	}
}

func runnerDecoratorA0R1[R0 any](r runner) DecoratorA0R1[R0] {
	return func(next NextA0R1[R0]) NextA0R1[R0] {
		return func(call CallA0R1) (error, R0) {
			if err := r(call.Key); err != nil {
				var zero R0
				return err, zero
			}
			return next(call)
		}
	}
}

func runnerDecoratorA1R1[A0, R0 any](r runner) DecoratorA1R1[A0, R0] {
	return func(next NextA1R1[A0, R0]) NextA1R1[A0, R0] {
		return func(call CallA1R1[A0]) (error, R0) {
			if err := r(call.Key); err != nil {
				var zero R0
				return err, zero
			}
			return next(call)
		}
	}
}

func runnerDecoratorA2R1[A0, A1, R0 any](r runner) DecoratorA2R1[A0, A1, R0] {
	return func(next NextA2R1[A0, A1, R0]) NextA2R1[A0, A1, R0] {
		return func(call CallA2R1[A0, A1]) (error, R0) {
			if err := r(call.Key); err != nil {
				var zero R0
				return err, zero
			}
			return next(call)
		}
	}
}

func BindA0R0(
	point string,
	base func() error,
) func() error {
	baseNext := NextA0R0(func(call CallA0R0) error {
		return base()
	})

	var cachedVersion atomic.Uint64
	var cachedNext atomic.Value
	cachedNext.Store(baseNext)

	return func() error {
		current := activeEngine.Load()
		if current == nil {
			return base()
		}

		version := current.version
		if cachedVersion.Load() != version {
			next := baseNext
			if r := current.Runner(point); r != nil {
				next = runnerDecoratorA0R0(r)(baseNext)
			}
			cachedNext.Store(next)
			cachedVersion.Store(version)
		}

		next := cachedNext.Load().(NextA0R0)
		return next(CallA0R0{})
	}
}

func BindA1R0[A0 any](
	point string,
	keyFn func(A0) string,
	base func(A0) error,
) func(A0) error {
	baseNext := NextA1R0[A0](func(call CallA1R0[A0]) error {
		return base(call.A0)
	})

	var cachedVersion atomic.Uint64
	var cachedNext atomic.Value
	cachedNext.Store(baseNext)

	return func(a0 A0) error {
		key := ""
		if keyFn != nil {
			key = keyFn(a0)
		}

		current := activeEngine.Load()
		if current == nil {
			return base(a0)
		}

		version := current.version
		if cachedVersion.Load() != version {
			next := baseNext
			if r := current.Runner(point); r != nil {
				next = runnerDecoratorA1R0[A0](r)(baseNext)
			}
			cachedNext.Store(next)
			cachedVersion.Store(version)
		}

		next := cachedNext.Load().(NextA1R0[A0])
		return next(CallA1R0[A0]{
			Key: key,
			A0:  a0,
		})
	}
}

func BindA2R0[A0, A1 any](
	point string,
	keyFn func(A0, A1) string,
	base func(A0, A1) error,
) func(A0, A1) error {
	baseNext := NextA2R0[A0, A1](func(call CallA2R0[A0, A1]) error {
		return base(call.A0, call.A1)
	})

	var cachedVersion atomic.Uint64
	var cachedNext atomic.Value
	cachedNext.Store(baseNext)

	return func(a0 A0, a1 A1) error {
		key := ""
		if keyFn != nil {
			key = keyFn(a0, a1)
		}

		current := activeEngine.Load()
		if current == nil {
			return base(a0, a1)
		}

		version := current.version
		if cachedVersion.Load() != version {
			next := baseNext
			if r := current.Runner(point); r != nil {
				next = runnerDecoratorA2R0[A0, A1](r)(baseNext)
			}
			cachedNext.Store(next)
			cachedVersion.Store(version)
		}

		next := cachedNext.Load().(NextA2R0[A0, A1])
		return next(CallA2R0[A0, A1]{
			Key: key,
			A0:  a0,
			A1:  a1,
		})
	}
}

func BindA0R1[R0 any](
	point string,
	base func() (error, R0),
) func() (error, R0) {
	baseNext := NextA0R1[R0](func(call CallA0R1) (error, R0) {
		return base()
	})

	var cachedVersion atomic.Uint64
	var cachedNext atomic.Value
	cachedNext.Store(baseNext)

	return func() (error, R0) {
		current := activeEngine.Load()
		if current == nil {
			return base()
		}

		version := current.version
		if cachedVersion.Load() != version {
			next := baseNext
			if r := current.Runner(point); r != nil {
				next = runnerDecoratorA0R1[R0](r)(baseNext)
			}
			cachedNext.Store(next)
			cachedVersion.Store(version)
		}

		next := cachedNext.Load().(NextA0R1[R0])
		return next(CallA0R1{})
	}
}

func BindA1R1[A0, R0 any](
	point string,
	keyFn func(A0) string,
	base func(A0) (error, R0),
) func(A0) (error, R0) {
	baseNext := NextA1R1[A0, R0](func(call CallA1R1[A0]) (error, R0) {
		return base(call.A0)
	})

	var cachedVersion atomic.Uint64
	var cachedNext atomic.Value
	cachedNext.Store(baseNext)

	return func(a0 A0) (error, R0) {
		key := ""
		if keyFn != nil {
			key = keyFn(a0)
		}

		current := activeEngine.Load()
		if current == nil {
			return base(a0)
		}

		version := current.version
		if cachedVersion.Load() != version {
			next := baseNext
			if r := current.Runner(point); r != nil {
				next = runnerDecoratorA1R1[A0, R0](r)(baseNext)
			}
			cachedNext.Store(next)
			cachedVersion.Store(version)
		}

		next := cachedNext.Load().(NextA1R1[A0, R0])
		return next(CallA1R1[A0]{
			Key: key,
			A0:  a0,
		})
	}
}

func BindA2R1[A0, A1, R0 any](
	point string,
	keyFn func(A0, A1) string,
	base func(A0, A1) (error, R0),
) func(A0, A1) (error, R0) {
	baseNext := NextA2R1[A0, A1, R0](func(call CallA2R1[A0, A1]) (error, R0) {
		return base(call.A0, call.A1)
	})

	var cachedVersion atomic.Uint64
	var cachedNext atomic.Value
	cachedNext.Store(baseNext)

	return func(a0 A0, a1 A1) (error, R0) {
		key := ""
		if keyFn != nil {
			key = keyFn(a0, a1)
		}

		current := activeEngine.Load()
		if current == nil {
			return base(a0, a1)
		}

		version := current.version
		if cachedVersion.Load() != version {
			next := baseNext
			if r := current.Runner(point); r != nil {
				next = runnerDecoratorA2R1[A0, A1, R0](r)(baseNext)
			}
			cachedNext.Store(next)
			cachedVersion.Store(version)
		}

		next := cachedNext.Load().(NextA2R1[A0, A1, R0])
		return next(CallA2R1[A0, A1]{
			Key: key,
			A0:  a0,
			A1:  a1,
		})
	}
}

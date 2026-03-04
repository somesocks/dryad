package diagnostics

import (
	"sync/atomic"
	"time"
)

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

type CallA3R0[A0, A1, A2 any] struct {
	Key string
	A0  A0
	A1  A1
	A2  A2
}

type NextA3R0[A0, A1, A2 any] func(CallA3R0[A0, A1, A2]) error
type RuleA3R0[A0, A1, A2 any] func(next NextA3R0[A0, A1, A2]) NextA3R0[A0, A1, A2]
type DecoratorA3R0[A0, A1, A2 any] = RuleA3R0[A0, A1, A2]

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

type CallA3R1[A0, A1, A2 any] struct {
	Key string
	A0  A0
	A1  A1
	A2  A2
}

type NextA3R1[A0, A1, A2, R0 any] func(CallA3R1[A0, A1, A2]) (error, R0)
type RuleA3R1[A0, A1, A2, R0 any] func(next NextA3R1[A0, A1, A2, R0]) NextA3R1[A0, A1, A2, R0]
type DecoratorA3R1[A0, A1, A2, R0 any] = RuleA3R1[A0, A1, A2, R0]

func ruleDecoratorA0R0(rule *compiledRule) DecoratorA0R0 {
	return func(next NextA0R0) NextA0R0 {
		return func(call CallA0R0) error {
			switch rule.action {
			case actionDelay:
				if rule.matches(call.Key) && rule.delay > 0 {
					time.Sleep(rule.delay)
				}
				return next(call)
			case actionError:
				hit := rule.matches(call.Key)
				if hit && !rule.postError {
					return rule.err
				}
				err := next(call)
				if err != nil {
					return err
				}
				if hit && rule.postError {
					return rule.err
				}
				return nil
			default:
				return next(call)
			}
		}
	}
}

func ruleDecoratorA1R0[A0 any](rule *compiledRule) DecoratorA1R0[A0] {
	return func(next NextA1R0[A0]) NextA1R0[A0] {
		return func(call CallA1R0[A0]) error {
			switch rule.action {
			case actionDelay:
				if rule.matches(call.Key) && rule.delay > 0 {
					time.Sleep(rule.delay)
				}
				return next(call)
			case actionError:
				hit := rule.matches(call.Key)
				if hit && !rule.postError {
					return rule.err
				}
				err := next(call)
				if err != nil {
					return err
				}
				if hit && rule.postError {
					return rule.err
				}
				return nil
			default:
				return next(call)
			}
		}
	}
}

func ruleDecoratorA2R0[A0, A1 any](rule *compiledRule) DecoratorA2R0[A0, A1] {
	return func(next NextA2R0[A0, A1]) NextA2R0[A0, A1] {
		return func(call CallA2R0[A0, A1]) error {
			switch rule.action {
			case actionDelay:
				if rule.matches(call.Key) && rule.delay > 0 {
					time.Sleep(rule.delay)
				}
				return next(call)
			case actionError:
				hit := rule.matches(call.Key)
				if hit && !rule.postError {
					return rule.err
				}
				err := next(call)
				if err != nil {
					return err
				}
				if hit && rule.postError {
					return rule.err
				}
				return nil
			default:
				return next(call)
			}
		}
	}
}

func ruleDecoratorA3R0[A0, A1, A2 any](rule *compiledRule) DecoratorA3R0[A0, A1, A2] {
	return func(next NextA3R0[A0, A1, A2]) NextA3R0[A0, A1, A2] {
		return func(call CallA3R0[A0, A1, A2]) error {
			switch rule.action {
			case actionDelay:
				if rule.matches(call.Key) && rule.delay > 0 {
					time.Sleep(rule.delay)
				}
				return next(call)
			case actionError:
				hit := rule.matches(call.Key)
				if hit && !rule.postError {
					return rule.err
				}
				err := next(call)
				if err != nil {
					return err
				}
				if hit && rule.postError {
					return rule.err
				}
				return nil
			default:
				return next(call)
			}
		}
	}
}

func ruleDecoratorA0R1[R0 any](rule *compiledRule) DecoratorA0R1[R0] {
	return func(next NextA0R1[R0]) NextA0R1[R0] {
		return func(call CallA0R1) (error, R0) {
			switch rule.action {
			case actionDelay:
				if rule.matches(call.Key) && rule.delay > 0 {
					time.Sleep(rule.delay)
				}
				return next(call)
			case actionError:
				hit := rule.matches(call.Key)
				if hit && !rule.postError {
					var zero R0
					return rule.err, zero
				}
				err, out := next(call)
				if err != nil {
					return err, out
				}
				if hit && rule.postError {
					return rule.err, out
				}
				return nil, out
			default:
				return next(call)
			}
		}
	}
}

func ruleDecoratorA1R1[A0, R0 any](rule *compiledRule) DecoratorA1R1[A0, R0] {
	return func(next NextA1R1[A0, R0]) NextA1R1[A0, R0] {
		return func(call CallA1R1[A0]) (error, R0) {
			switch rule.action {
			case actionDelay:
				if rule.matches(call.Key) && rule.delay > 0 {
					time.Sleep(rule.delay)
				}
				return next(call)
			case actionError:
				hit := rule.matches(call.Key)
				if hit && !rule.postError {
					var zero R0
					return rule.err, zero
				}
				err, out := next(call)
				if err != nil {
					return err, out
				}
				if hit && rule.postError {
					return rule.err, out
				}
				return nil, out
			default:
				return next(call)
			}
		}
	}
}

func ruleDecoratorA2R1[A0, A1, R0 any](rule *compiledRule) DecoratorA2R1[A0, A1, R0] {
	return func(next NextA2R1[A0, A1, R0]) NextA2R1[A0, A1, R0] {
		return func(call CallA2R1[A0, A1]) (error, R0) {
			switch rule.action {
			case actionDelay:
				if rule.matches(call.Key) && rule.delay > 0 {
					time.Sleep(rule.delay)
				}
				return next(call)
			case actionError:
				hit := rule.matches(call.Key)
				if hit && !rule.postError {
					var zero R0
					return rule.err, zero
				}
				err, out := next(call)
				if err != nil {
					return err, out
				}
				if hit && rule.postError {
					return rule.err, out
				}
				return nil, out
			default:
				return next(call)
			}
		}
	}
}

func ruleDecoratorA3R1[A0, A1, A2, R0 any](rule *compiledRule) DecoratorA3R1[A0, A1, A2, R0] {
	return func(next NextA3R1[A0, A1, A2, R0]) NextA3R1[A0, A1, A2, R0] {
		return func(call CallA3R1[A0, A1, A2]) (error, R0) {
			switch rule.action {
			case actionDelay:
				if rule.matches(call.Key) && rule.delay > 0 {
					time.Sleep(rule.delay)
				}
				return next(call)
			case actionError:
				hit := rule.matches(call.Key)
				if hit && !rule.postError {
					var zero R0
					return rule.err, zero
				}
				err, out := next(call)
				if err != nil {
					return err, out
				}
				if hit && rule.postError {
					return rule.err, out
				}
				return nil, out
			default:
				return next(call)
			}
		}
	}
}

func buildA0R0Next(current *engine, point string, base func() error) NextA0R0 {
	next := NextA0R0(func(call CallA0R0) error {
		return base()
	})
	rules := current.Rules(point)
	for i := len(rules) - 1; i >= 0; i-- {
		next = ruleDecoratorA0R0(rules[i])(next)
	}
	return next
}

func buildA1R0Next[A0 any](current *engine, point string, base func(A0) error) NextA1R0[A0] {
	next := NextA1R0[A0](func(call CallA1R0[A0]) error {
		return base(call.A0)
	})
	rules := current.Rules(point)
	for i := len(rules) - 1; i >= 0; i-- {
		next = ruleDecoratorA1R0[A0](rules[i])(next)
	}
	return next
}

func buildA2R0Next[A0, A1 any](current *engine, point string, base func(A0, A1) error) NextA2R0[A0, A1] {
	next := NextA2R0[A0, A1](func(call CallA2R0[A0, A1]) error {
		return base(call.A0, call.A1)
	})
	rules := current.Rules(point)
	for i := len(rules) - 1; i >= 0; i-- {
		next = ruleDecoratorA2R0[A0, A1](rules[i])(next)
	}
	return next
}

func buildA3R0Next[A0, A1, A2 any](current *engine, point string, base func(A0, A1, A2) error) NextA3R0[A0, A1, A2] {
	next := NextA3R0[A0, A1, A2](func(call CallA3R0[A0, A1, A2]) error {
		return base(call.A0, call.A1, call.A2)
	})
	rules := current.Rules(point)
	for i := len(rules) - 1; i >= 0; i-- {
		next = ruleDecoratorA3R0[A0, A1, A2](rules[i])(next)
	}
	return next
}

func buildA0R1Next[R0 any](current *engine, point string, base func() (error, R0)) NextA0R1[R0] {
	next := NextA0R1[R0](func(call CallA0R1) (error, R0) {
		return base()
	})
	rules := current.Rules(point)
	for i := len(rules) - 1; i >= 0; i-- {
		next = ruleDecoratorA0R1[R0](rules[i])(next)
	}
	return next
}

func buildA1R1Next[A0, R0 any](current *engine, point string, base func(A0) (error, R0)) NextA1R1[A0, R0] {
	next := NextA1R1[A0, R0](func(call CallA1R1[A0]) (error, R0) {
		return base(call.A0)
	})
	rules := current.Rules(point)
	for i := len(rules) - 1; i >= 0; i-- {
		next = ruleDecoratorA1R1[A0, R0](rules[i])(next)
	}
	return next
}

func buildA2R1Next[A0, A1, R0 any](current *engine, point string, base func(A0, A1) (error, R0)) NextA2R1[A0, A1, R0] {
	next := NextA2R1[A0, A1, R0](func(call CallA2R1[A0, A1]) (error, R0) {
		return base(call.A0, call.A1)
	})
	rules := current.Rules(point)
	for i := len(rules) - 1; i >= 0; i-- {
		next = ruleDecoratorA2R1[A0, A1, R0](rules[i])(next)
	}
	return next
}

func buildA3R1Next[A0, A1, A2, R0 any](current *engine, point string, base func(A0, A1, A2) (error, R0)) NextA3R1[A0, A1, A2, R0] {
	next := NextA3R1[A0, A1, A2, R0](func(call CallA3R1[A0, A1, A2]) (error, R0) {
		return base(call.A0, call.A1, call.A2)
	})
	rules := current.Rules(point)
	for i := len(rules) - 1; i >= 0; i-- {
		next = ruleDecoratorA3R1[A0, A1, A2, R0](rules[i])(next)
	}
	return next
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
			cachedNext.Store(buildA0R0Next(current, point, base))
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
			cachedNext.Store(buildA1R0Next[A0](current, point, base))
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
			cachedNext.Store(buildA2R0Next[A0, A1](current, point, base))
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

func BindA3R0[A0, A1, A2 any](
	point string,
	keyFn func(A0, A1, A2) string,
	base func(A0, A1, A2) error,
) func(A0, A1, A2) error {
	baseNext := NextA3R0[A0, A1, A2](func(call CallA3R0[A0, A1, A2]) error {
		return base(call.A0, call.A1, call.A2)
	})

	var cachedVersion atomic.Uint64
	var cachedNext atomic.Value
	cachedNext.Store(baseNext)

	return func(a0 A0, a1 A1, a2 A2) error {
		key := ""
		if keyFn != nil {
			key = keyFn(a0, a1, a2)
		}

		current := activeEngine.Load()
		if current == nil {
			return base(a0, a1, a2)
		}

		version := current.version
		if cachedVersion.Load() != version {
			cachedNext.Store(buildA3R0Next[A0, A1, A2](current, point, base))
			cachedVersion.Store(version)
		}

		next := cachedNext.Load().(NextA3R0[A0, A1, A2])
		return next(CallA3R0[A0, A1, A2]{
			Key: key,
			A0:  a0,
			A1:  a1,
			A2:  a2,
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
			cachedNext.Store(buildA0R1Next[R0](current, point, base))
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
			cachedNext.Store(buildA1R1Next[A0, R0](current, point, base))
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
			cachedNext.Store(buildA2R1Next[A0, A1, R0](current, point, base))
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

func BindA3R1[A0, A1, A2, R0 any](
	point string,
	keyFn func(A0, A1, A2) string,
	base func(A0, A1, A2) (error, R0),
) func(A0, A1, A2) (error, R0) {
	baseNext := NextA3R1[A0, A1, A2, R0](func(call CallA3R1[A0, A1, A2]) (error, R0) {
		return base(call.A0, call.A1, call.A2)
	})

	var cachedVersion atomic.Uint64
	var cachedNext atomic.Value
	cachedNext.Store(baseNext)

	return func(a0 A0, a1 A1, a2 A2) (error, R0) {
		key := ""
		if keyFn != nil {
			key = keyFn(a0, a1, a2)
		}

		current := activeEngine.Load()
		if current == nil {
			return base(a0, a1, a2)
		}

		version := current.version
		if cachedVersion.Load() != version {
			cachedNext.Store(buildA3R1Next[A0, A1, A2, R0](current, point, base))
			cachedVersion.Store(version)
		}

		next := cachedNext.Load().(NextA3R1[A0, A1, A2, R0])
		return next(CallA3R1[A0, A1, A2]{
			Key: key,
			A0:  a0,
			A1:  a1,
			A2:  a2,
		})
	}
}

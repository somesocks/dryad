package diagnostics

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type engine struct {
	version uint64
	rules   map[string][]*compiledRule
}

func (e *engine) Rules(point string) []*compiledRule {
	return e.rules[point]
}

type keyMatcherKind int

const (
	keyMatcherAny keyMatcherKind = iota
	keyMatcherExact
	keyMatcherPrefix
)

type keyMatcher struct {
	kind   keyMatcherKind
	value  string
	prefix string
}

func (m keyMatcher) Matches(key string) bool {
	switch m.kind {
	case keyMatcherAny:
		return true
	case keyMatcherExact:
		return key == m.value
	case keyMatcherPrefix:
		return strings.HasPrefix(key, m.prefix)
	default:
		return false
	}
}

type whenMode int

const (
	whenOncePerKey whenMode = iota
	whenFirstNPerKey
	whenEveryN
	whenPercent
)

type actionType int

const (
	actionError actionType = iota
	actionDelay
)

type compiledRule struct {
	id        string
	op        string
	matcher   keyMatcher
	when      whenMode
	n         uint64
	percent   uint64
	maxHits   int64
	hitCount  atomic.Int64
	counter   atomic.Uint64
	perKeyMu  sync.Mutex
	perKey    map[uint64]uint64
	action    actionType
	postError bool
	delay     time.Duration
	err       error
	seed      uint64
}

func compileConfig(cfg Config) (*engine, error) {
	if cfg.Version != 1 {
		return nil, fmt.Errorf("unsupported diagnostics version %d", cfg.Version)
	}

	rulesByPoint := map[string][]*compiledRule{}

	for idx, rule := range cfg.Rules {
		if rule.Enabled != nil && !*rule.Enabled {
			continue
		}

		compiled, op, err := compileRule(cfg.Seed, idx, rule)
		if err != nil {
			return nil, err
		}
		rulesByPoint[op] = append(rulesByPoint[op], compiled)
	}

	return &engine{rules: rulesByPoint}, nil
}

func compileRule(seed int64, index int, rule RuleConfig) (*compiledRule, string, error) {
	op := strings.TrimSpace(rule.Op)
	if op == "" {
		return nil, "", fmt.Errorf("diagnostics rule missing op")
	}

	matcher, err := compileKeyMatcher(rule.Key)
	if err != nil {
		return nil, "", fmt.Errorf("diagnostics rule %q: %w", rule.ID, err)
	}

	compiled := &compiledRule{
		id:      rule.ID,
		op:      op,
		matcher: matcher,
		maxHits: rule.MaxHits,
		seed:    mix64(uint64(seed) ^ uint64(index+1)),
	}

	if err := compileWhen(compiled, rule.When, rule.ID); err != nil {
		return nil, "", err
	}
	if err := compileAction(compiled, rule.Action, rule.ID); err != nil {
		return nil, "", err
	}

	return compiled, op, nil
}

func compileKeyMatcher(raw string) (keyMatcher, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return keyMatcher{}, fmt.Errorf("diagnostics rule key is required")
	}

	if raw == "*" {
		return keyMatcher{kind: keyMatcherAny}, nil
	}

	if strings.HasPrefix(raw, "prefix:") {
		prefix := strings.TrimPrefix(raw, "prefix:")
		if prefix == "" {
			return keyMatcher{}, fmt.Errorf("diagnostics key prefix is empty")
		}
		return keyMatcher{kind: keyMatcherPrefix, prefix: prefix}, nil
	}

	return keyMatcher{kind: keyMatcherExact, value: raw}, nil
}

func compileWhen(out *compiledRule, when WhenConfig, id string) error {
	count := when.Count

	switch when.Mode {
	case "once_per_key":
		out.when = whenOncePerKey
		out.perKey = map[uint64]uint64{}
	case "first_n_per_key":
		if count <= 0 {
			return fmt.Errorf("diagnostics rule %q: when.count must be > 0", id)
		}
		out.when = whenFirstNPerKey
		out.n = uint64(count)
		out.perKey = map[uint64]uint64{}
	case "every_n":
		if count <= 0 {
			return fmt.Errorf("diagnostics rule %q: when.count must be > 0", id)
		}
		out.when = whenEveryN
		out.n = uint64(count)
	case "percent":
		if when.Percent < 0 || when.Percent > 100 {
			return fmt.Errorf("diagnostics rule %q: when.percent must be in [0,100]", id)
		}
		out.when = whenPercent
		out.percent = uint64(math.Round((when.Percent / 100.0) * 1_000_000.0))
	default:
		return fmt.Errorf("diagnostics rule %q: unsupported when.mode %q", id, when.Mode)
	}

	return nil
}

func compileAction(out *compiledRule, action ActionConfig, id string) error {
	switch action.Type {
	case "error":
		errValue, err := parseErrorName(action.Error)
		if err != nil {
			return fmt.Errorf("diagnostics rule %q: %w", id, err)
		}
		post, err := parseErrorPhase(action.Phase)
		if err != nil {
			return fmt.Errorf("diagnostics rule %q: %w", id, err)
		}
		out.action = actionError
		out.postError = post
		out.err = errValue
	case "delay":
		if strings.TrimSpace(action.Phase) != "" {
			return fmt.Errorf("diagnostics rule %q: action.phase is only supported for action.type=\"error\"", id)
		}
		if action.DelayMS < 0 {
			return fmt.Errorf("diagnostics rule %q: action.delay_ms must be >= 0", id)
		}
		out.action = actionDelay
		out.delay = time.Duration(action.DelayMS) * time.Millisecond
	default:
		return fmt.Errorf("diagnostics rule %q: unsupported action.type %q", id, action.Type)
	}

	return nil
}

func parseErrorPhase(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "pre":
		return false, nil
	case "post":
		return true, nil
	default:
		return false, fmt.Errorf("unsupported action.phase %q", raw)
	}
}

func parseErrorName(name string) (error, error) {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "EMLINK":
		return syscall.EMLINK, nil
	case "EXDEV":
		return syscall.EXDEV, nil
	case "EIO":
		return syscall.EIO, nil
	case "ETIMEDOUT":
		return syscall.ETIMEDOUT, nil
	default:
		return nil, fmt.Errorf("unsupported error %q", name)
	}
}

func (rule *compiledRule) matches(key string) bool {
	if !rule.matcher.Matches(key) {
		return false
	}
	if !rule.whenMatches(key) {
		return false
	}
	if !rule.consumeHit() {
		return false
	}
	return true
}

func (rule *compiledRule) whenMatches(key string) bool {
	switch rule.when {
	case whenOncePerKey:
		keyHash := hashString64(key)
		rule.perKeyMu.Lock()
		defer rule.perKeyMu.Unlock()
		if rule.perKey[keyHash] > 0 {
			return false
		}
		rule.perKey[keyHash] = 1
		return true

	case whenFirstNPerKey:
		keyHash := hashString64(key)
		rule.perKeyMu.Lock()
		defer rule.perKeyMu.Unlock()
		current := rule.perKey[keyHash]
		if current >= rule.n {
			return false
		}
		rule.perKey[keyHash] = current + 1
		return true

	case whenEveryN:
		count := rule.counter.Add(1)
		return count%rule.n == 0

	case whenPercent:
		count := rule.counter.Add(1)
		roll := mix64(rule.seed ^ hashString64(key) ^ count)
		return (roll % 1_000_000) < rule.percent

	default:
		return false
	}
}

func (rule *compiledRule) consumeHit() bool {
	if rule.maxHits <= 0 {
		return true
	}

	for {
		current := rule.hitCount.Load()
		if current >= rule.maxHits {
			return false
		}
		if rule.hitCount.CompareAndSwap(current, current+1) {
			return true
		}
	}
}

func hashString64(s string) uint64 {
	const offset = 1469598103934665603
	const prime = 1099511628211

	var h uint64 = offset
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime
	}
	return h
}

func mix64(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}

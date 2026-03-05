package diagnostics

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type engine struct {
	version uint64
	rules   map[string][]*compiledRule
	metrics map[string]*compiledMetricsRule
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
	whenBeforeN whenMode = iota
	whenAfterN
	whenBeforeNPerKey
	whenAfterNPerKey
	whenEveryNPerKey
	whenEveryN
)

type actionType int

const (
	actionError actionType = iota
	actionDelay
	actionMetrics
)

type metricsOutputKind int

const (
	metricsOutputStdout metricsOutputKind = iota
	metricsOutputStderr
)

type compiledMetricsRule struct {
	id            string
	op            string
	output        metricsOutputKind
	captureCalls  bool
	captureErrors bool
	captureTiming bool
	sampleEvery   uint64
	stats         atomic.Pointer[pointStats]
}

type compiledRule struct {
	id        string
	op        string
	matcher   keyMatcher
	when      whenMode
	n         uint64
	limit     int64
	hitCount  atomic.Int64
	counter   atomic.Uint64
	perKeyMu  sync.Mutex
	perKey    map[uint64]uint64
	action    actionType
	postError bool
	delay     time.Duration
	err       error
	metric    *compiledMetricsRule
}

func compileConfig(cfg Config) (*engine, error) {
	if cfg.Version != 1 {
		return nil, fmt.Errorf("unsupported diagnostics version %d", cfg.Version)
	}

	rulesByPoint := map[string][]*compiledRule{}
	metricsByID := map[string]*compiledMetricsRule{}

	for idx, rule := range cfg.Rules {
		if rule.Enabled != nil && !*rule.Enabled {
			continue
		}

		compiled, op, err := compileRule(idx, rule)
		if err != nil {
			return nil, err
		}
		rulesByPoint[op] = append(rulesByPoint[op], compiled)
		if compiled.metric != nil {
			if _, exists := metricsByID[compiled.metric.id]; exists {
				return nil, fmt.Errorf("duplicate diagnostics metrics rule id %q", compiled.metric.id)
			}
			compiled.metric.stats.Store(pointStatsFor(compiled.metric.id))
			metricsByID[compiled.metric.id] = compiled.metric
		}
	}

	return &engine{
		rules:   rulesByPoint,
		metrics: metricsByID,
	}, nil
}

func compileRule(index int, rule RuleConfig) (*compiledRule, string, error) {
	ruleID := fallbackRuleID("rule", index, rule.ID)

	op := strings.TrimSpace(rule.Op)
	if op == "" {
		return nil, "", fmt.Errorf("diagnostics rule %q: missing op", ruleID)
	}

	matcher, err := compileKeyMatcher(rule.Key)
	if err != nil {
		return nil, "", fmt.Errorf("diagnostics rule %q: %w", ruleID, err)
	}

	compiled := &compiledRule{
		id:      ruleID,
		op:      op,
		matcher: matcher,
		limit:   rule.When.Limit,
	}

	if err := compileWhen(compiled, rule.When, ruleID); err != nil {
		return nil, "", err
	}
	if err := compileAction(compiled, rule.Action, ruleID); err != nil {
		return nil, "", err
	}

	return compiled, op, nil
}

func fallbackRuleID(prefix string, index int, configured string) string {
	id := strings.TrimSpace(configured)
	if id != "" {
		return id
	}
	return fmt.Sprintf("%s-%d", prefix, index+1)
}

func boolOrDefault(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}

func compileMetricsAction(ruleID string, when whenMode, n uint64, op string, action ActionConfig) (*compiledMetricsRule, error) {
	if strings.TrimSpace(action.Phase) != "" {
		return nil, fmt.Errorf("diagnostics rule %q: action.phase is only supported for action.type=\"error\"", ruleID)
	}
	if strings.TrimSpace(action.Error) != "" {
		return nil, fmt.Errorf("diagnostics rule %q: action.error is only supported for action.type=\"error\"", ruleID)
	}
	if action.DelayMS != 0 {
		return nil, fmt.Errorf("diagnostics rule %q: action.delay_ms is only supported for action.type=\"delay\"", ruleID)
	}

	output, err := parseMetricsOutput(action.Output)
	if err != nil {
		return nil, fmt.Errorf("diagnostics rule %q: %w", ruleID, err)
	}

	captureCalls := boolOrDefault(action.Capture.Calls, true)
	captureErrors := boolOrDefault(action.Capture.Errors, true)
	captureTiming := boolOrDefault(action.Capture.Timing, true)
	if !captureCalls && !captureErrors && !captureTiming {
		return nil, fmt.Errorf("diagnostics rule %q: capture must enable at least one of calls, errors, timing", ruleID)
	}

	sampleEvery := uint64(1)
	if when == whenEveryN {
		sampleEvery = n
	}

	return &compiledMetricsRule{
		id:            ruleID,
		op:            op,
		output:        output,
		captureCalls:  captureCalls,
		captureErrors: captureErrors,
		captureTiming: captureTiming,
		sampleEvery:   sampleEvery,
	}, nil
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
	x := when.X

	switch when.Mode {
	case "before_x":
		if x <= 0 {
			return fmt.Errorf("diagnostics rule %q: when.x must be > 0", id)
		}
		out.when = whenBeforeN
		out.n = uint64(x)
	case "after_x":
		if x <= 0 {
			return fmt.Errorf("diagnostics rule %q: when.x must be > 0", id)
		}
		out.when = whenAfterN
		out.n = uint64(x)
	case "before_x_per_key":
		if x <= 0 {
			return fmt.Errorf("diagnostics rule %q: when.x must be > 0", id)
		}
		out.when = whenBeforeNPerKey
		out.n = uint64(x)
		out.perKey = map[uint64]uint64{}
	case "after_x_per_key":
		if x <= 0 {
			return fmt.Errorf("diagnostics rule %q: when.x must be > 0", id)
		}
		out.when = whenAfterNPerKey
		out.n = uint64(x)
		out.perKey = map[uint64]uint64{}
	case "every_x_per_key":
		if x <= 0 {
			return fmt.Errorf("diagnostics rule %q: when.x must be > 0", id)
		}
		out.when = whenEveryNPerKey
		out.n = uint64(x)
		out.perKey = map[uint64]uint64{}
	case "every_x":
		if x <= 0 {
			return fmt.Errorf("diagnostics rule %q: when.x must be > 0", id)
		}
		out.when = whenEveryN
		out.n = uint64(x)
	default:
		return fmt.Errorf("diagnostics rule %q: unsupported when.mode %q", id, when.Mode)
	}

	return nil
}

func compileAction(out *compiledRule, action ActionConfig, id string) error {
	switch action.Type {
	case "error":
		if strings.TrimSpace(action.Output) != "" || hasActionCaptureConfig(action.Capture) {
			return fmt.Errorf("diagnostics rule %q: action.output and action.capture are only supported for action.type=\"metrics\"", id)
		}
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
		if strings.TrimSpace(action.Output) != "" || hasActionCaptureConfig(action.Capture) {
			return fmt.Errorf("diagnostics rule %q: action.output and action.capture are only supported for action.type=\"metrics\"", id)
		}
		if strings.TrimSpace(action.Phase) != "" {
			return fmt.Errorf("diagnostics rule %q: action.phase is only supported for action.type=\"error\"", id)
		}
		if action.DelayMS < 0 {
			return fmt.Errorf("diagnostics rule %q: action.delay_ms must be >= 0", id)
		}
		out.action = actionDelay
		out.delay = time.Duration(action.DelayMS) * time.Millisecond
	case "metrics":
		metric, err := compileMetricsAction(id, out.when, out.n, out.op, action)
		if err != nil {
			return err
		}
		out.action = actionMetrics
		out.metric = metric
	default:
		return fmt.Errorf("diagnostics rule %q: unsupported action.type %q", id, action.Type)
	}

	return nil
}

func hasActionCaptureConfig(capture MetricsCaptureConfig) bool {
	return capture.Calls != nil || capture.Errors != nil || capture.Timing != nil
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

func parseMetricsOutput(raw string) (metricsOutputKind, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "stderr":
		return metricsOutputStderr, nil
	case "stdout":
		return metricsOutputStdout, nil
	default:
		return metricsOutputStderr, fmt.Errorf("unsupported metrics output %q", raw)
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
	case whenBeforeN:
		count := rule.counter.Add(1)
		return count <= rule.n

	case whenAfterN:
		count := rule.counter.Add(1)
		return count > rule.n

	case whenBeforeNPerKey:
		keyHash := hashString64(key)
		rule.perKeyMu.Lock()
		defer rule.perKeyMu.Unlock()
		current := rule.perKey[keyHash]
		if current >= rule.n {
			return false
		}
		rule.perKey[keyHash] = current + 1
		return true

	case whenAfterNPerKey:
		keyHash := hashString64(key)
		rule.perKeyMu.Lock()
		defer rule.perKeyMu.Unlock()
		current := rule.perKey[keyHash] + 1
		rule.perKey[keyHash] = current
		return current > rule.n

	case whenEveryNPerKey:
		keyHash := hashString64(key)
		rule.perKeyMu.Lock()
		defer rule.perKeyMu.Unlock()
		current := rule.perKey[keyHash] + 1
		rule.perKey[keyHash] = current
		return current%rule.n == 0

	case whenEveryN:
		count := rule.counter.Add(1)
		return count%rule.n == 0

	default:
		return false
	}
}

func (rule *compiledRule) consumeHit() bool {
	if rule.limit <= 0 {
		return true
	}

	for {
		current := rule.hitCount.Load()
		if current >= rule.limit {
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

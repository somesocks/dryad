package diagnostics

import (
	"strconv"
	"testing"
)

var benchSinkErr error
var benchSinkInt int

func benchBaseA2R0(a0 string, a1 string) error {
	if len(a0)+len(a1) == -1 {
		benchSinkInt++
	}
	return nil
}

func benchMetricsRule(capture MetricsCaptureConfig) RuleConfig {
	return benchMetricsRuleEveryN(capture, 1)
}

func benchMetricsRuleEveryN(capture MetricsCaptureConfig, n int64) RuleConfig {
	return benchMetricsRuleWithWhen(capture, WhenConfig{Mode: "every_x", X: n})
}

func benchMetricsRuleWithWhen(capture MetricsCaptureConfig, when WhenConfig) RuleConfig {
	return RuleConfig{
		ID:   "m",
		Op:   "os.link",
		Key:  "*",
		When: when,
		Action: ActionConfig{
			Type:    "metrics",
			Capture: capture,
		},
	}
}

func benchStaticKeys(count int) []string {
	keys := make([]string, count)
	for i := range count {
		keys[i] = "k-" + strconv.Itoa(i)
	}
	return keys
}

func BenchmarkBindA2R0_Direct(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchSinkErr = benchBaseA2R0("a", "b")
	}
}

func BenchmarkBindA2R0_Disabled(b *testing.B) {
	Disable()
	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledNoRules(b *testing.B) {
	if err := SetupFromConfig(Config{Version: 1, Rules: nil}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Disable)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledPreErrorNoHit(b *testing.B) {
	if err := SetupFromConfig(Config{
		Version: 1,
		Seed:    1,
		Rules: []RuleConfig{{
			ID:   "e",
			Op:   "os.link",
			Key:  "zzz",
			When: WhenConfig{Mode: "every_x", X: 1},
			Action: ActionConfig{
				Type:  "error",
				Error: "EMLINK",
			},
		}},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Disable)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledPostErrorNoHit(b *testing.B) {
	if err := SetupFromConfig(Config{
		Version: 1,
		Seed:    1,
		Rules: []RuleConfig{{
			ID:   "e",
			Op:   "os.link",
			Key:  "zzz",
			When: WhenConfig{Mode: "every_x", X: 1},
			Action: ActionConfig{
				Type:  "error",
				Phase: "post",
				Error: "EMLINK",
			},
		}},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Disable)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledMetricsAll(b *testing.B) {
	Reset()
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRule(MetricsCaptureConfig{}),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledMetricsNoTiming(b *testing.B) {
	Reset()
	disabled := false
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRule(MetricsCaptureConfig{
				Timing: &disabled,
			}),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledMetricsCallsOnly(b *testing.B) {
	Reset()
	disabled := false
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRule(MetricsCaptureConfig{
				Calls:  nil,
				Errors: &disabled,
				Timing: &disabled,
			}),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledMetricsErrorsOnly(b *testing.B) {
	Reset()
	enabled := true
	disabled := false
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRule(MetricsCaptureConfig{
				Calls:  &disabled,
				Errors: &enabled,
				Timing: &disabled,
			}),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledMetricsCallsAndErrorsOnly(b *testing.B) {
	Reset()
	enabled := true
	disabled := false
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRule(MetricsCaptureConfig{
				Calls:  &enabled,
				Errors: &enabled,
				Timing: &disabled,
			}),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledMetricsTimingOnly(b *testing.B) {
	Reset()
	disabled := false
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRule(MetricsCaptureConfig{
				Calls:  &disabled,
				Errors: &disabled,
				Timing: nil,
			}),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledMetricsNoTimingSample50(b *testing.B) {
	Reset()
	disabled := false
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRuleEveryN(MetricsCaptureConfig{
				Timing: &disabled,
			}, 2),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledMetricsAllSample50(b *testing.B) {
	Reset()
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRuleEveryN(MetricsCaptureConfig{}, 2),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledMetricsAllSample1(b *testing.B) {
	Reset()
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRuleEveryN(MetricsCaptureConfig{}, 128),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a0 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "b")
	}
}

func BenchmarkBindA2R0_EnabledMetricsBeforeXPerKeyCount1NoTiming_SameKey(b *testing.B) {
	Reset()
	disabled := false
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRuleWithWhen(
				MetricsCaptureConfig{Timing: &disabled},
				WhenConfig{Mode: "before_x_per_key", X: 1},
			),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a1 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "shared")
	}
}

func BenchmarkBindA2R0_EnabledMetricsBeforeXPerKeyCount4NoTiming_SameKey(b *testing.B) {
	Reset()
	disabled := false
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRuleWithWhen(
				MetricsCaptureConfig{Timing: &disabled},
				WhenConfig{Mode: "before_x_per_key", X: 4},
			),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a1 }, benchBaseA2R0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", "shared")
	}
}

func BenchmarkBindA2R0_EnabledMetricsBeforeXPerKeyCount1NoTiming_UniqueKeys(b *testing.B) {
	Reset()
	disabled := false
	const keyCount = 1 << 16
	keys := benchStaticKeys(keyCount)
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRuleWithWhen(
				MetricsCaptureConfig{Timing: &disabled},
				WhenConfig{Mode: "before_x_per_key", X: 1},
			),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a1 }, benchBaseA2R0)
	mask := keyCount - 1
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", keys[i&mask])
	}
}

func BenchmarkBindA2R0_EnabledMetricsBeforeXPerKeyCount4NoTiming_UniqueKeys(b *testing.B) {
	Reset()
	disabled := false
	const keyCount = 1 << 16
	keys := benchStaticKeys(keyCount)
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRuleWithWhen(
				MetricsCaptureConfig{Timing: &disabled},
				WhenConfig{Mode: "before_x_per_key", X: 4},
			),
		},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(Reset)

	bound := BindA2R0("os.link", func(a0 string, a1 string) string { return a1 }, benchBaseA2R0)
	mask := keyCount - 1
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkErr = bound("a", keys[i&mask])
	}
}

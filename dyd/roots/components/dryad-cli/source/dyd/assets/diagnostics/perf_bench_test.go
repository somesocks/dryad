package diagnostics

import "testing"

var benchSinkErr error
var benchSinkInt int

func benchBaseA2R0(a0 string, a1 string) error {
	if len(a0)+len(a1) == -1 {
		benchSinkInt++
	}
	return nil
}

func benchMetricsRule(capture MetricsCaptureConfig) RuleConfig {
	return RuleConfig{
		ID:   "m",
		Op:   "os.link",
		Key:  "*",
		When: WhenConfig{Mode: "every_n", Count: 1},
		Action: ActionConfig{
			Type:    "metrics",
			Capture: capture,
		},
	}
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
			When: WhenConfig{Mode: "every_n", Count: 1},
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
			When: WhenConfig{Mode: "every_n", Count: 1},
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
	samplePercent := 50.0
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRule(MetricsCaptureConfig{
				Timing:        &disabled,
				SamplePercent: &samplePercent,
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

func BenchmarkBindA2R0_EnabledMetricsAllSample50(b *testing.B) {
	Reset()
	samplePercent := 50.0
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRule(MetricsCaptureConfig{
				SamplePercent: &samplePercent,
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

func BenchmarkBindA2R0_EnabledMetricsAllSample1(b *testing.B) {
	Reset()
	samplePercent := 1.0
	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			benchMetricsRule(MetricsCaptureConfig{
				SamplePercent: &samplePercent,
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

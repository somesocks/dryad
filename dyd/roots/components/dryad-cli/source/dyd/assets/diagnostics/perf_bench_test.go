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
		Metrics: []MetricsRuleConfig{
			{
				ID: "m",
				Op: "os.link",
			},
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
		Metrics: []MetricsRuleConfig{
			{
				ID: "m",
				Op: "os.link",
				Capture: MetricsCaptureConfig{
					Timing: &disabled,
				},
			},
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
		Metrics: []MetricsRuleConfig{
			{
				ID: "m",
				Op: "os.link",
				Capture: MetricsCaptureConfig{
					Calls:  nil,
					Errors: &disabled,
					Timing: &disabled,
				},
			},
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
		Metrics: []MetricsRuleConfig{
			{
				ID: "m",
				Op: "os.link",
				Capture: MetricsCaptureConfig{
					Calls:  &disabled,
					Errors: &enabled,
					Timing: &disabled,
				},
			},
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
		Metrics: []MetricsRuleConfig{
			{
				ID: "m",
				Op: "os.link",
				Capture: MetricsCaptureConfig{
					Calls:  &enabled,
					Errors: &enabled,
					Timing: &disabled,
				},
			},
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
		Metrics: []MetricsRuleConfig{
			{
				ID: "m",
				Op: "os.link",
				Capture: MetricsCaptureConfig{
					Calls:  &disabled,
					Errors: &disabled,
					Timing: nil,
				},
			},
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

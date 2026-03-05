package diagnostics

import (
	"errors"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

func boolRef(v bool) *bool {
	return &v
}

func metricsRule(id string, op string, output string, capture MetricsCaptureConfig) RuleConfig {
	return metricsRuleWithWhen(id, op, output, WhenConfig{Mode: "every_x", X: 1}, capture)
}

func metricsRuleWithWhen(id string, op string, output string, when WhenConfig, capture MetricsCaptureConfig) RuleConfig {
	return RuleConfig{
		ID:   id,
		Op:   op,
		Key:  "*",
		When: when,
		Action: ActionConfig{
			Type:    "metrics",
			Output:  output,
			Capture: capture,
		},
	}
}

func TestMetricsSnapshot_BinderCountsAndErrors(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRule("m-bind", "metrics.bind", "stderr", MetricsCaptureConfig{}),
		},
	}); err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	expectedErr := errors.New("boom")
	bound := BindA1R0(
		"metrics.bind",
		nil,
		func(v int) error {
			if v == 2 {
				return expectedErr
			}
			return nil
		},
	)

	if err := bound(1); err != nil {
		t.Fatalf("expected nil on first call, got %v", err)
	}
	if err := bound(2); !errors.Is(err, expectedErr) {
		t.Fatalf("expected boom on second call, got %v", err)
	}

	snapshot := MetricsSnapshot()
	stats, ok := snapshot["m-bind"]
	if !ok {
		t.Fatalf("missing metrics rule m-bind")
	}
	if stats.Calls != 2 {
		t.Fatalf("expected calls=2, got %d", stats.Calls)
	}
	if stats.Errors != 1 {
		t.Fatalf("expected errors=1, got %d", stats.Errors)
	}
	if stats.AvgNanos != stats.TotalNanos/stats.Calls {
		t.Fatalf("expected avg=%d, got %d", stats.TotalNanos/stats.Calls, stats.AvgNanos)
	}
	if stats.MinNanos > stats.MaxNanos {
		t.Fatalf("expected min <= max, got min=%d max=%d", stats.MinNanos, stats.MaxNanos)
	}
}

func TestMetricsSnapshot_ApplyCounts(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	err := SetupFromConfig(Config{
		Version: 1,
		Seed:    1,
		Rules: []RuleConfig{
			metricsRule("m-apply", "metrics.apply", "stderr", MetricsCaptureConfig{}),
			{
				ID:   "inject",
				Op:   "metrics.apply",
				Key:  "*",
				When: WhenConfig{Mode: "every_x", X: 1},
				Action: ActionConfig{
					Type:  "error",
					Error: "EIO",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	if err := Apply("metrics.apply", "a"); !errors.Is(err, syscall.EIO) {
		t.Fatalf("expected EIO with diagnostics enabled, got %v", err)
	}

	snapshot := MetricsSnapshot()
	stats, ok := snapshot["m-apply"]
	if !ok {
		t.Fatalf("missing metrics rule m-apply")
	}
	if stats.Calls != 1 {
		t.Fatalf("expected calls=1, got %d", stats.Calls)
	}
	if stats.Errors != 1 {
		t.Fatalf("expected errors=1, got %d", stats.Errors)
	}
}

func TestResetMetrics_ClearsState(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRule("m-reset", "metrics.reset", "stderr", MetricsCaptureConfig{}),
		},
	}); err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	bound := BindA0R0(
		"metrics.reset",
		func() error { return nil },
	)
	if err := bound(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if len(MetricsSnapshot()) == 0 {
		t.Fatalf("expected metrics before reset")
	}

	Reset()
	if len(MetricsSnapshot()) != 0 {
		t.Fatalf("expected empty metrics after reset")
	}
	if activeEngine.Load() != nil {
		t.Fatalf("expected active engine to be nil after reset")
	}
}

func TestResetMetrics_RebindsActiveMetricStats(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRule("m-reset-active", "metrics.reset_active", "stderr", MetricsCaptureConfig{}),
		},
	}); err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	bound := BindA0R0(
		"metrics.reset_active",
		func() error { return nil },
	)
	if err := bound(); err != nil {
		t.Fatalf("expected nil on first call, got %v", err)
	}

	before := MetricsSnapshot()["m-reset-active"]
	if before.Calls != 1 {
		t.Fatalf("expected calls=1 before reset, got %d", before.Calls)
	}

	ResetMetrics()
	resetSnap := MetricsSnapshot()["m-reset-active"]
	if resetSnap.Calls != 0 || resetSnap.Errors != 0 || resetSnap.TotalNanos != 0 {
		t.Fatalf("expected zeroed stats immediately after reset, got %+v", resetSnap)
	}

	if err := bound(); err != nil {
		t.Fatalf("expected nil on second call, got %v", err)
	}

	after := MetricsSnapshot()["m-reset-active"]
	if after.Calls != 1 {
		t.Fatalf("expected calls=1 after reset and one new call, got %d", after.Calls)
	}
}

func TestEmitMetricsOnExit_UsesConfiguredStream(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRule("m-out", "metrics.output", "stdout", MetricsCaptureConfig{}),
		},
	}); err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	bound := BindA0R0(
		"metrics.output",
		func() error { return nil },
	)
	if err := bound(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
		_ = r.Close()
	}()

	if err := EmitMetricsOnExit(); err != nil {
		t.Fatalf("emit metrics: %v", err)
	}
	_ = w.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout capture: %v", err)
	}

	if !strings.Contains(string(data), `"point":"metrics.output"`) {
		t.Fatalf("expected output to include metrics point, got: %s", string(data))
	}
	if !strings.Contains(string(data), `"sample_every":1`) {
		t.Fatalf("expected output to include sample_every=1, got: %s", string(data))
	}
}

func TestSetupFromConfig_MetricsCaptureRejectsAllDisabled(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRule("m-disabled", "metrics.none", "", MetricsCaptureConfig{
				Calls:  boolRef(false),
				Errors: boolRef(false),
				Timing: boolRef(false),
			}),
		},
	})
	if err == nil {
		t.Fatalf("expected setup to fail when all metrics capture flags are disabled")
	}
}

func TestSetupFromConfig_GeneratesMetricsRuleIDWhenMissing(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRule("", "metrics.generated_id", "", MetricsCaptureConfig{
				Calls:  boolRef(false),
				Errors: boolRef(false),
				Timing: boolRef(false),
			}),
		},
	})
	if err == nil {
		t.Fatalf("expected setup to fail when all metrics capture flags are disabled")
	}
	if !strings.Contains(err.Error(), `diagnostics rule "rule-1":`) {
		t.Fatalf("expected generated metrics rule id in error, got: %v", err)
	}
}

func TestMetricsSnapshot_EveryNControlsSampling(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRuleWithWhen(
				"m-sample-every2",
				"metrics.sample_all",
				"",
				WhenConfig{Mode: "every_x", X: 2},
				MetricsCaptureConfig{},
			),
		},
	}); err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	bound := BindA0R0(
		"metrics.sample_all",
		func() error {
			time.Sleep(200 * time.Microsecond)
			return syscall.EIO
		},
	)

	for i := 0; i < 4; i++ {
		if err := bound(); err != syscall.EIO {
			t.Fatalf("expected base EIO, got %v", err)
		}
	}

	stats := MetricsSnapshot()["m-sample-every2"]
	if stats.Calls != 2 {
		t.Fatalf("expected sampled calls=2, got %d", stats.Calls)
	}
	if stats.Errors != 2 {
		t.Fatalf("expected sampled errors=2, got %d", stats.Errors)
	}
	if stats.TotalNanos == 0 || stats.AvgNanos == 0 {
		t.Fatalf("expected sampled timing metrics to be non-zero, got %+v", stats)
	}
}

func TestEmitMetricsOnExit_IncludesSampleEveryFromWhenEveryN(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRuleWithWhen(
				"m-sample-every",
				"metrics.sample_every",
				"",
				WhenConfig{Mode: "every_x", X: 2},
				MetricsCaptureConfig{Timing: boolRef(false)},
			),
		},
	}); err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	bound := BindA0R0(
		"metrics.sample_every",
		func() error { return nil },
	)
	if err := bound(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
		_ = r.Close()
	}()

	if err := EmitMetricsOnExit(); err != nil {
		t.Fatalf("emit metrics: %v", err)
	}
	_ = w.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stderr capture: %v", err)
	}

	out := string(data)
	if !strings.Contains(out, `"sample_every":2`) {
		t.Fatalf("expected sample_every=2 in output, got: %s", out)
	}
}

func TestMetricsSnapshot_CaptureTimingDisabled(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRule("m-no-timing", "metrics.no_timing", "", MetricsCaptureConfig{
				Timing: boolRef(false),
			}),
		},
	}); err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	bound := BindA0R0(
		"metrics.no_timing",
		func() error { return nil },
	)
	if err := bound(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	stats := MetricsSnapshot()["m-no-timing"]
	if stats.Calls != 1 {
		t.Fatalf("expected calls=1, got %d", stats.Calls)
	}
	if stats.TotalNanos != 0 || stats.MinNanos != 0 || stats.MaxNanos != 0 || stats.AvgNanos != 0 {
		t.Fatalf("expected timing stats to stay 0 when timing capture is disabled, got %+v", stats)
	}
}

func TestMetricsSnapshot_CaptureErrorsDisabled(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRule("m-no-errors", "metrics.no_errors", "", MetricsCaptureConfig{
				Errors: boolRef(false),
			}),
		},
	}); err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	expectedErr := errors.New("boom")
	bound := BindA0R0(
		"metrics.no_errors",
		func() error { return expectedErr },
	)

	if err := bound(); !errors.Is(err, expectedErr) {
		t.Fatalf("expected error from base call, got %v", err)
	}

	stats := MetricsSnapshot()["m-no-errors"]
	if stats.Calls != 1 {
		t.Fatalf("expected calls=1, got %d", stats.Calls)
	}
	if stats.Errors != 0 {
		t.Fatalf("expected errors=0 when errors capture is disabled, got %d", stats.Errors)
	}
}

func TestMetricsSnapshot_CaptureCallsDisabled(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRule("m-no-calls", "metrics.no_calls", "", MetricsCaptureConfig{
				Calls: boolRef(false),
			}),
		},
	}); err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	bound := BindA0R0(
		"metrics.no_calls",
		func() error {
			time.Sleep(1 * time.Millisecond)
			return nil
		},
	)
	if err := bound(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	stats := MetricsSnapshot()["m-no-calls"]
	if stats.Calls != 0 {
		t.Fatalf("expected calls=0 when calls capture is disabled, got %d", stats.Calls)
	}
	if stats.TotalNanos == 0 || stats.AvgNanos == 0 {
		t.Fatalf("expected non-zero timing stats when timing capture is enabled, got %+v", stats)
	}
}

func TestMetricsRuleOrdering_PreErrorCanBypassLaterMetrics(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			{
				ID:   "pre-error-first",
				Op:   "metrics.order",
				Key:  "*",
				When: WhenConfig{Mode: "every_x", X: 1},
				Action: ActionConfig{
					Type:  "error",
					Error: "EIO",
				},
			},
			metricsRule("m-later", "metrics.order", "", MetricsCaptureConfig{}),
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	if err := Apply("metrics.order", "x"); !errors.Is(err, syscall.EIO) {
		t.Fatalf("expected EIO, got %v", err)
	}

	snapshot := MetricsSnapshot()
	if stats, ok := snapshot["m-later"]; ok && stats.Calls != 0 {
		t.Fatalf("expected later metrics rule to observe zero calls, got %+v", stats)
	}
}

func TestMetricsRuleOrdering_MetricsBeforePreErrorObservesError(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			metricsRule("m-first", "metrics.order", "", MetricsCaptureConfig{}),
			{
				ID:   "pre-error-second",
				Op:   "metrics.order",
				Key:  "*",
				When: WhenConfig{Mode: "every_x", X: 1},
				Action: ActionConfig{
					Type:  "error",
					Error: "EIO",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	if err := Apply("metrics.order", "x"); !errors.Is(err, syscall.EIO) {
		t.Fatalf("expected EIO, got %v", err)
	}

	stats, ok := MetricsSnapshot()["m-first"]
	if !ok {
		t.Fatalf("missing metrics rule m-first")
	}
	if stats.Calls != 1 || stats.Errors != 1 {
		t.Fatalf("expected calls=1 errors=1, got %+v", stats)
	}
}

func TestMetricsSnapshot_MultipleMetricsRulesOnSameOp(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			{
				ID:   "m-a-prefix",
				Op:   "metrics.multi",
				Key:  "prefix:prefix:a",
				When: WhenConfig{Mode: "every_x", X: 1},
				Action: ActionConfig{
					Type:    "metrics",
					Capture: MetricsCaptureConfig{Timing: boolRef(false)},
				},
			},
			{
				ID:   "m-b-prefix",
				Op:   "metrics.multi",
				Key:  "prefix:prefix:b",
				When: WhenConfig{Mode: "every_x", X: 1},
				Action: ActionConfig{
					Type:    "metrics",
					Capture: MetricsCaptureConfig{Timing: boolRef(false)},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	if err := Apply("metrics.multi", "prefix:a"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := Apply("metrics.multi", "prefix:a"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := Apply("metrics.multi", "prefix:b"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	snapshot := MetricsSnapshot()
	statsA, ok := snapshot["m-a-prefix"]
	if !ok {
		t.Fatalf("missing metrics rule m-a-prefix")
	}
	statsB, ok := snapshot["m-b-prefix"]
	if !ok {
		t.Fatalf("missing metrics rule m-b-prefix")
	}

	if statsA.Calls != 2 || statsA.Errors != 0 {
		t.Fatalf("expected m-a-prefix calls=2 errors=0, got %+v", statsA)
	}
	if statsB.Calls != 1 || statsB.Errors != 0 {
		t.Fatalf("expected m-b-prefix calls=1 errors=0, got %+v", statsB)
	}
}

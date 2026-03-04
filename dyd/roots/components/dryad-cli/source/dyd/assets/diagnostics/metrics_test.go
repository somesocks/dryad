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

func floatRef(v float64) *float64 {
	return &v
}

func TestMetricsSnapshot_BinderCountsAndErrors(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Metrics: []MetricsRuleConfig{
			{ID: "m-bind", Op: "metrics.bind", Output: "stderr"},
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
	stats, ok := snapshot["metrics.bind"]
	if !ok {
		t.Fatalf("missing metrics point metrics.bind")
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
			{
				ID:   "inject",
				Op:   "metrics.apply",
				Key:  "*",
				When: WhenConfig{Mode: "every_n", Count: 1},
				Action: ActionConfig{
					Type:  "error",
					Error: "EIO",
				},
			},
		},
		Metrics: []MetricsRuleConfig{
			{ID: "m-apply", Op: "metrics.apply", Output: "stderr"},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	if err := Apply("metrics.apply", "a"); !errors.Is(err, syscall.EIO) {
		t.Fatalf("expected EIO with diagnostics enabled, got %v", err)
	}

	snapshot := MetricsSnapshot()
	stats, ok := snapshot["metrics.apply"]
	if !ok {
		t.Fatalf("missing metrics point metrics.apply")
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
		Metrics: []MetricsRuleConfig{
			{ID: "m-reset", Op: "metrics.reset", Output: "stderr"},
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
		Metrics: []MetricsRuleConfig{
			{ID: "m-reset-active", Op: "metrics.reset_active", Output: "stderr"},
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

	before := MetricsSnapshot()["metrics.reset_active"]
	if before.Calls != 1 {
		t.Fatalf("expected calls=1 before reset, got %d", before.Calls)
	}

	ResetMetrics()
	resetSnap := MetricsSnapshot()["metrics.reset_active"]
	if resetSnap.Calls != 0 || resetSnap.Errors != 0 || resetSnap.TotalNanos != 0 {
		t.Fatalf("expected zeroed stats immediately after reset, got %+v", resetSnap)
	}

	if err := bound(); err != nil {
		t.Fatalf("expected nil on second call, got %v", err)
	}

	after := MetricsSnapshot()["metrics.reset_active"]
	if after.Calls != 1 {
		t.Fatalf("expected calls=1 after reset and one new call, got %d", after.Calls)
	}
}

func TestEmitMetricsOnExit_UsesConfiguredStream(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Metrics: []MetricsRuleConfig{
			{ID: "m-out", Op: "metrics.output", Output: "stdout"},
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
		Metrics: []MetricsRuleConfig{
			{
				ID: "m-disabled",
				Op: "metrics.none",
				Capture: MetricsCaptureConfig{
					Calls:  boolRef(false),
					Errors: boolRef(false),
					Timing: boolRef(false),
				},
			},
		},
	})
	if err == nil {
		t.Fatalf("expected setup to fail when all metrics capture flags are disabled")
	}
}

func TestSetupFromConfig_MetricsSamplePercentRejectsOutOfRange(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	err := SetupFromConfig(Config{
		Version: 1,
		Metrics: []MetricsRuleConfig{
			{
				ID: "m-bad-sample",
				Op: "metrics.bad_sample",
				Capture: MetricsCaptureConfig{
					SamplePercent: floatRef(0),
				},
			},
		},
	})
	if err == nil {
		t.Fatalf("expected setup to fail for invalid sample percent")
	}
}

func TestMetricsSnapshot_SamplePercentAffectsAllMetrics(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Metrics: []MetricsRuleConfig{
			{
				ID: "m-sample-all",
				Op: "metrics.sample_all",
				Capture: MetricsCaptureConfig{
					SamplePercent: floatRef(50),
				},
			},
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

	stats := MetricsSnapshot()["metrics.sample_all"]
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

func TestMetricsSnapshot_SamplePercentRoundsToPowerOfTwoRate(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Metrics: []MetricsRuleConfig{
			{
				ID: "m-sample-round",
				Op: "metrics.sample_round",
				Capture: MetricsCaptureConfig{
					SamplePercent: floatRef(30),
					Timing:        boolRef(false),
				},
			},
		},
	}); err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	bound := BindA0R0(
		"metrics.sample_round",
		func() error { return nil },
	)

	for i := 0; i < 8; i++ {
		if err := bound(); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	}

	stats := MetricsSnapshot()["metrics.sample_round"]
	// 30% rounds to nearest power-of-two capture rate: 25% (1-in-4).
	if stats.Calls != 2 {
		t.Fatalf("expected sampled calls=2 for 8 invocations at rounded 25%% rate, got %d", stats.Calls)
	}
}

func TestEmitMetricsOnExit_IncludesSampleEvery(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	if err := SetupFromConfig(Config{
		Version: 1,
		Metrics: []MetricsRuleConfig{
			{
				ID: "m-sample-every",
				Op: "metrics.sample_every",
				Capture: MetricsCaptureConfig{
					SamplePercent: floatRef(50),
					Timing:        boolRef(false),
				},
			},
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
		Metrics: []MetricsRuleConfig{
			{
				ID: "m-no-timing",
				Op: "metrics.no_timing",
				Capture: MetricsCaptureConfig{
					Timing: boolRef(false),
				},
			},
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

	stats := MetricsSnapshot()["metrics.no_timing"]
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
		Metrics: []MetricsRuleConfig{
			{
				ID: "m-no-errors",
				Op: "metrics.no_errors",
				Capture: MetricsCaptureConfig{
					Errors: boolRef(false),
				},
			},
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

	stats := MetricsSnapshot()["metrics.no_errors"]
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
		Metrics: []MetricsRuleConfig{
			{
				ID: "m-no-calls",
				Op: "metrics.no_calls",
				Capture: MetricsCaptureConfig{
					Calls: boolRef(false),
				},
			},
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

	stats := MetricsSnapshot()["metrics.no_calls"]
	if stats.Calls != 0 {
		t.Fatalf("expected calls=0 when calls capture is disabled, got %d", stats.Calls)
	}
	if stats.TotalNanos == 0 || stats.AvgNanos == 0 {
		t.Fatalf("expected non-zero timing stats when timing capture is enabled, got %+v", stats)
	}
}

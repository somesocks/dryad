package diagnostics

import (
	"errors"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
)

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
}

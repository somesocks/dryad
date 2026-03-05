package diagnostics

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestSetupFromEnv_JSON_BeforeXPerKeyCount1Error(t *testing.T) {
	t.Setenv(
		EnvVar,
		`json:{"version":1,"seed":123,"rules":[{"id":"x","op":"heap.link_to_stem","key":"*","when":{"mode":"before_x_per_key","x":1},"action":{"type":"error","error":"EMLINK"}}]}`,
	)

	if err := SetupFromEnv(); err != nil {
		t.Fatalf("SetupFromEnv failed: %v", err)
	}
	t.Cleanup(Disable)

	err := Apply("heap.link_to_stem", "alpha")
	if !errors.Is(err, syscall.EMLINK) {
		t.Fatalf("expected EMLINK on first key hit, got: %v", err)
	}

	err = Apply("heap.link_to_stem", "alpha")
	if err != nil {
		t.Fatalf("expected nil on second key hit, got: %v", err)
	}

	err = Apply("heap.link_to_stem", "beta")
	if !errors.Is(err, syscall.EMLINK) {
		t.Fatalf("expected EMLINK on new key hit, got: %v", err)
	}
}

func TestSetupFromConfig_BeforeX_Global(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			{
				ID:   "before-global",
				Op:   "heap.link_to_stem",
				Key:  "*",
				When: WhenConfig{Mode: "before_x", X: 2},
				Action: ActionConfig{
					Type:  "error",
					Error: "EMLINK",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	if err := Apply("heap.link_to_stem", "k"); !errors.Is(err, syscall.EMLINK) {
		t.Fatalf("expected EMLINK on first global hit, got: %v", err)
	}
	if err := Apply("heap.link_to_stem", "k"); !errors.Is(err, syscall.EMLINK) {
		t.Fatalf("expected EMLINK on second global hit, got: %v", err)
	}
	if err := Apply("heap.link_to_stem", "k"); err != nil {
		t.Fatalf("expected nil after before_x window closes, got: %v", err)
	}
}

func TestSetupFromConfig_AfterX_Global(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			{
				ID:   "after-global",
				Op:   "heap.link_to_stem",
				Key:  "*",
				When: WhenConfig{Mode: "after_x", X: 2},
				Action: ActionConfig{
					Type:  "error",
					Error: "EMLINK",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	if err := Apply("heap.link_to_stem", "k"); err != nil {
		t.Fatalf("expected nil before after_x threshold, got: %v", err)
	}
	if err := Apply("heap.link_to_stem", "k"); err != nil {
		t.Fatalf("expected nil before after_x threshold, got: %v", err)
	}
	if err := Apply("heap.link_to_stem", "k"); !errors.Is(err, syscall.EMLINK) {
		t.Fatalf("expected EMLINK once after_x threshold is crossed, got: %v", err)
	}
}

func TestSetupFromConfig_AfterXPerKey(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			{
				ID:   "after-per-key",
				Op:   "heap.link_to_stem",
				Key:  "*",
				When: WhenConfig{Mode: "after_x_per_key", X: 1},
				Action: ActionConfig{
					Type:  "error",
					Error: "EMLINK",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	if err := Apply("heap.link_to_stem", "alpha"); err != nil {
		t.Fatalf("expected nil before alpha threshold, got: %v", err)
	}
	if err := Apply("heap.link_to_stem", "alpha"); !errors.Is(err, syscall.EMLINK) {
		t.Fatalf("expected EMLINK after alpha threshold, got: %v", err)
	}
	if err := Apply("heap.link_to_stem", "beta"); err != nil {
		t.Fatalf("expected nil before beta threshold, got: %v", err)
	}
	if err := Apply("heap.link_to_stem", "beta"); !errors.Is(err, syscall.EMLINK) {
		t.Fatalf("expected EMLINK after beta threshold, got: %v", err)
	}
}

func TestSetupFromEnv_File_EveryNDelay(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "diag.yaml")
	content := []byte(`
version: 1
seed: 7
rules:
  - id: delay-every-2
    op: heap.link_to_sprout
    key: "*"
    when:
      mode: every_x
      x: 2
    action:
      type: delay
      delay_ms: 5
`)

	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv(EnvVar, "file:"+configPath)

	if err := SetupFromEnv(); err != nil {
		t.Fatalf("SetupFromEnv failed: %v", err)
	}
	t.Cleanup(Disable)

	start := time.Now()
	if err := Apply("heap.link_to_sprout", "x"); err != nil {
		t.Fatalf("expected no error on first delay call, got %v", err)
	}
	firstElapsed := time.Since(start)

	start = time.Now()
	if err := Apply("heap.link_to_sprout", "x"); err != nil {
		t.Fatalf("expected no error on second delay call, got %v", err)
	}
	secondElapsed := time.Since(start)

	if secondElapsed < 4*time.Millisecond {
		t.Fatalf("expected second call to include delay, first=%v second=%v", firstElapsed, secondElapsed)
	}
}

func TestSetupFromEnv_InvalidPrefix(t *testing.T) {
	t.Setenv(EnvVar, "oops")
	if err := SetupFromEnv(); err == nil {
		t.Fatal("expected error for invalid DYD_DIAG prefix")
	}
}

func TestSetupFromConfig_ErrorPhaseRejectsUnknownValue(t *testing.T) {
	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			{
				ID:   "bad-phase",
				Op:   "os.link",
				Key:  "*",
				When: WhenConfig{Mode: "every_x", X: 1},
				Action: ActionConfig{
					Type:  "error",
					Phase: "later",
					Error: "EMLINK",
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected compile error for unsupported action.phase")
	}
	if !strings.Contains(err.Error(), `unsupported action.phase`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetupFromConfig_DelayRejectsPhase(t *testing.T) {
	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			{
				ID:   "bad-delay-phase",
				Op:   "os.link",
				Key:  "*",
				When: WhenConfig{Mode: "every_x", X: 1},
				Action: ActionConfig{
					Type:    "delay",
					Phase:   "post",
					DelayMS: 1,
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected compile error for delay action.phase")
	}
	if !strings.Contains(err.Error(), `action.phase is only supported`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetupFromConfig_GeneratesRuleIDWhenMissing(t *testing.T) {
	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			{
				Op:   "os.link",
				Key:  "*",
				When: WhenConfig{Mode: "every_x", X: 0},
				Action: ActionConfig{
					Type:  "error",
					Error: "EMLINK",
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected compile error for invalid when.x")
	}
	if !strings.Contains(err.Error(), `diagnostics rule "rule-1":`) {
		t.Fatalf("expected generated rule id in error, got: %v", err)
	}
}

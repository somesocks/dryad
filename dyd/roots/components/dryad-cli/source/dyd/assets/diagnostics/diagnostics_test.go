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

func TestSetupFromEnv_JSON_OncePerKeyError(t *testing.T) {
	t.Setenv(
		EnvVar,
		`json:{"version":1,"seed":123,"rules":[{"id":"x","op":"heap.link_to_stem","key":"*","when":{"mode":"once_per_key"},"action":{"type":"error","error":"EMLINK"}}]}`,
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
      mode: every_n
      count: 2
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
		t.Fatal("expected error for invalid DYD_DG prefix")
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
				When: WhenConfig{Mode: "every_n", Count: 1},
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
				When: WhenConfig{Mode: "every_n", Count: 1},
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

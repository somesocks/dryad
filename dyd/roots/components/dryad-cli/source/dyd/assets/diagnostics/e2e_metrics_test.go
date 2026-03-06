package diagnostics

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestE2E_MetricsFromEnv_EmitOnExit(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	configPath := filepath.Join(t.TempDir(), "diag.yaml")
	configBody := []byte(`
version: 1
seed: 1
rules:
  - id: m-e2e
    op: e2e.point
    key: "*"
    when:
      mode: every_x
      x: 1
    action:
      type: metrics
      output: stderr
  - id: inject
    op: e2e.point
    key: "*"
    when:
      mode: every_x
      x: 1
    action:
      type: error
      error: EIO
`)
	if err := os.WriteFile(configPath, configBody, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv(EnvVar, "file:"+configPath)
	if err := SetupFromEnv(); err != nil {
		t.Fatalf("setup from env: %v", err)
	}

	bound := BindA0R0(
		"e2e.point",
		func() error { return nil },
	)
	if err := bound(); err != syscall.EIO {
		t.Fatalf("expected injected EIO, got %v", err)
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
	if !strings.Contains(out, `"point":"e2e.point"`) {
		t.Fatalf("missing point in output: %s", out)
	}
	if !strings.Contains(out, `"calls":1`) {
		t.Fatalf("missing calls in output: %s", out)
	}
	if !strings.Contains(out, `"errors":1`) {
		t.Fatalf("missing errors in output: %s", out)
	}
	if !strings.Contains(out, `"sample_every":1`) {
		t.Fatalf("missing sample_every in output: %s", out)
	}
}

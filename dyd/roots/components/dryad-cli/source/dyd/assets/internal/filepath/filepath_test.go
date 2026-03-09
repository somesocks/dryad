package filepath

import (
	"dryad/diagnostics"
	"errors"
	"os"
	"syscall"
	"testing"
)

func TestAbs_UsesDiagnostics(t *testing.T) {
	diagnostics.Disable()
	t.Cleanup(diagnostics.Disable)

	err := diagnostics.SetupFromConfig(diagnostics.Config{
		Version: 1,
		Rules: []diagnostics.RuleConfig{
			{
				ID:   "abs-error",
				Op:   "filepath.abs",
				Key:  "*",
				When: diagnostics.WhenConfig{Mode: "every_x", X: 1},
				Action: diagnostics.ActionConfig{
					Type:  "error",
					Error: "EIO",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	_, err = Abs(".")
	if !errors.Is(err, syscall.EIO) {
		t.Fatalf("expected EIO, got %v", err)
	}
}

func TestEvalSymlinks_UsesDiagnostics(t *testing.T) {
	diagnostics.Disable()
	t.Cleanup(diagnostics.Disable)

	err := diagnostics.SetupFromConfig(diagnostics.Config{
		Version: 1,
		Rules: []diagnostics.RuleConfig{
			{
				ID:   "eval-error",
				Op:   "filepath.eval_symlinks",
				Key:  "*",
				When: diagnostics.WhenConfig{Mode: "every_x", X: 1},
				Action: diagnostics.ActionConfig{
					Type:  "error",
					Error: "EIO",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	tempDir := t.TempDir()
	linkPath := Join(tempDir, "link")
	err = os.Symlink(tempDir, linkPath)
	if err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, err = EvalSymlinks(linkPath)
	if !errors.Is(err, syscall.EIO) {
		t.Fatalf("expected EIO, got %v", err)
	}
}

func TestRel_UsesDiagnostics(t *testing.T) {
	diagnostics.Disable()
	t.Cleanup(diagnostics.Disable)

	err := diagnostics.SetupFromConfig(diagnostics.Config{
		Version: 1,
		Rules: []diagnostics.RuleConfig{
			{
				ID:   "rel-error",
				Op:   "filepath.rel",
				Key:  "prefix:/tmp/target",
				When: diagnostics.WhenConfig{Mode: "every_x", X: 1},
				Action: diagnostics.ActionConfig{
					Type:  "error",
					Error: "EIO",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	_, err = Rel("/tmp/base", "/tmp/target/child")
	if !errors.Is(err, syscall.EIO) {
		t.Fatalf("expected EIO, got %v", err)
	}
}

package diagnostics

import (
	"errors"
	"sync/atomic"
	"syscall"
	"testing"
)

func TestBindA2R0_PicksUpEngineChangesWithoutRebinding(t *testing.T) {
	Disable()
	t.Cleanup(Disable)

	var calls atomic.Int64
	base := func(a0 string, a1 string) error {
		calls.Add(1)
		return nil
	}

	bound := BindA2R0(
		"os.link",
		func(a0 string, _ string) string {
			return a0
		},
		base,
	)

	// With diagnostics disabled, bound wrapper should call through.
	if err := bound("alpha", "dest"); err != nil {
		t.Fatalf("expected nil before setup, got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected base call count 1 before setup, got %d", calls.Load())
	}

	// Enable diagnostics with a deterministic error for all keys.
	err := SetupFromConfig(Config{
		Version: 1,
		Seed:    1,
		Rules: []RuleConfig{
			{
				ID:   "inject-error",
				Op:   "os.link",
				Key:  "*",
				When: WhenConfig{Mode: "every_x", X: 1},
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

	// Same previously-bound closure should now return injected error.
	err = bound("alpha", "dest")
	if !errors.Is(err, syscall.EMLINK) {
		t.Fatalf("expected injected EMLINK after setup, got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected no extra base call after injected error, got %d", calls.Load())
	}

	// Replace engine with no rules; same closure should fall through again.
	err = SetupFromConfig(Config{
		Version: 1,
		Seed:    2,
		Rules:   []RuleConfig{},
	})
	if err != nil {
		t.Fatalf("setup diagnostics (clear rules): %v", err)
	}

	err = bound("alpha", "dest")
	if err != nil {
		t.Fatalf("expected nil after clearing rules, got %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected base call count 2 after clearing rules, got %d", calls.Load())
	}
}

func TestBindA2R1_ReturnsZeroResultOnInjectedError(t *testing.T) {
	Disable()
	t.Cleanup(Disable)

	base := func(a0 string, a1 string) (error, int) {
		return nil, 42
	}

	bound := BindA2R1(
		"query.fetch",
		func(a0 string, a1 string) string {
			return a0 + ":" + a1
		},
		base,
	)

	err := SetupFromConfig(Config{
		Version: 1,
		Seed:    9,
		Rules: []RuleConfig{
			{
				ID:   "inject-error",
				Op:   "query.fetch",
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

	gotErr, gotVal := bound("a", "b")
	if !errors.Is(gotErr, syscall.EIO) {
		t.Fatalf("expected EIO, got %v", gotErr)
	}
	if gotVal != 0 {
		t.Fatalf("expected zero result on injected error, got %d", gotVal)
	}
}

func TestBindA2R0_PostErrorRunsAfterBase(t *testing.T) {
	Disable()
	t.Cleanup(Disable)

	var calls atomic.Int64
	base := func(a0 string, a1 string) error {
		calls.Add(1)
		return nil
	}

	bound := BindA2R0(
		"os.link",
		func(a0 string, _ string) string {
			return a0
		},
		base,
	)

	err := SetupFromConfig(Config{
		Version: 1,
		Seed:    5,
		Rules: []RuleConfig{
			{
				ID:   "inject-post-error",
				Op:   "os.link",
				Key:  "*",
				When: WhenConfig{Mode: "every_x", X: 1},
				Action: ActionConfig{
					Type:  "error",
					Phase: "post",
					Error: "EMLINK",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	err = bound("alpha", "dest")
	if !errors.Is(err, syscall.EMLINK) {
		t.Fatalf("expected post EMLINK, got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected base to run before post error, got calls=%d", calls.Load())
	}
}

func TestBindA2R1_PostErrorPreservesResult(t *testing.T) {
	Disable()
	t.Cleanup(Disable)

	base := func(a0 string, a1 string) (error, int) {
		return nil, 42
	}

	bound := BindA2R1(
		"query.fetch",
		func(a0 string, a1 string) string {
			return a0 + ":" + a1
		},
		base,
	)

	err := SetupFromConfig(Config{
		Version: 1,
		Seed:    7,
		Rules: []RuleConfig{
			{
				ID:   "inject-post-error",
				Op:   "query.fetch",
				Key:  "*",
				When: WhenConfig{Mode: "every_x", X: 1},
				Action: ActionConfig{
					Type:  "error",
					Phase: "post",
					Error: "EIO",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	gotErr, gotVal := bound("a", "b")
	if !errors.Is(gotErr, syscall.EIO) {
		t.Fatalf("expected post EIO, got %v", gotErr)
	}
	if gotVal != 42 {
		t.Fatalf("expected base result to be preserved on post error, got %d", gotVal)
	}
}

func TestBindA3R0_PicksUpEngineChangesWithoutRebinding(t *testing.T) {
	Disable()
	t.Cleanup(Disable)

	var calls atomic.Int64
	base := func(a0 string, a1 int, a2 bool) error {
		calls.Add(1)
		return nil
	}

	bound := BindA3R0(
		"os.open_file",
		func(a0 string, _ int, _ bool) string {
			return a0
		},
		base,
	)

	if err := bound("alpha", 1, true); err != nil {
		t.Fatalf("expected nil before setup, got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected base call count 1 before setup, got %d", calls.Load())
	}

	err := SetupFromConfig(Config{
		Version: 1,
		Seed:    11,
		Rules: []RuleConfig{
			{
				ID:   "inject-error",
				Op:   "os.open_file",
				Key:  "*",
				When: WhenConfig{Mode: "every_x", X: 1},
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

	err = bound("alpha", 1, true)
	if !errors.Is(err, syscall.EMLINK) {
		t.Fatalf("expected injected EMLINK after setup, got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected no extra base call after injected error, got %d", calls.Load())
	}
}

func TestBindA3R1_PostErrorPreservesResult(t *testing.T) {
	Disable()
	t.Cleanup(Disable)

	base := func(a0 string, a1 int, a2 bool) (error, int) {
		return nil, 42
	}

	bound := BindA3R1(
		"query.open_file",
		func(a0 string, _ int, _ bool) string {
			return a0
		},
		base,
	)

	err := SetupFromConfig(Config{
		Version: 1,
		Seed:    13,
		Rules: []RuleConfig{
			{
				ID:   "inject-post-error",
				Op:   "query.open_file",
				Key:  "*",
				When: WhenConfig{Mode: "every_x", X: 1},
				Action: ActionConfig{
					Type:  "error",
					Phase: "post",
					Error: "EIO",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("setup diagnostics: %v", err)
	}

	gotErr, gotVal := bound("alpha", 1, true)
	if !errors.Is(gotErr, syscall.EIO) {
		t.Fatalf("expected post EIO, got %v", gotErr)
	}
	if gotVal != 42 {
		t.Fatalf("expected base result to be preserved on post error, got %d", gotVal)
	}
}

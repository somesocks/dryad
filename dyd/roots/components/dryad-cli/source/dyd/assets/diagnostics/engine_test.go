package diagnostics

import (
	"strings"
	"testing"
)

func TestSetupFromConfig_WhenPercentUnsupported(t *testing.T) {
	err := SetupFromConfig(Config{
		Version: 1,
		Rules: []RuleConfig{
			{
				ID:   "p",
				Op:   "os.link",
				Key:  "*",
				When: WhenConfig{Mode: "percent"},
				Action: ActionConfig{
					Type:  "error",
					Error: "EIO",
				},
			},
		},
	})
	if err == nil {
		t.Fatalf("expected compile error for unsupported when.mode percent")
	}
	if !strings.Contains(err.Error(), `unsupported when.mode "percent"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

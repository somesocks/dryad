package diagnostics

import (
	"strings"
	"testing"
)

func TestSetupFromConfig_WhenPercentUnsupported(t *testing.T) {
	tests := []string{"percent", "once_per_key"}
	for _, mode := range tests {
		mode := mode
		t.Run(mode, func(t *testing.T) {
			err := SetupFromConfig(Config{
				Version: 1,
				Rules: []RuleConfig{
					{
						ID:   "p",
						Op:   "os.link",
						Key:  "*",
						When: WhenConfig{Mode: mode},
						Action: ActionConfig{
							Type:  "error",
							Error: "EIO",
						},
					},
				},
			})
			if err == nil {
				t.Fatalf("expected compile error for unsupported when.mode %q", mode)
			}
			expected := `unsupported when.mode "` + mode + `"`
			if !strings.Contains(err.Error(), expected) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

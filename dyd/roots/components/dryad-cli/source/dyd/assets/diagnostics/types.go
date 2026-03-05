package diagnostics

const (
	// EnvVar controls diagnostics configuration.
	// Supported values:
	// - file:/path/to/config.yaml
	// - json:{"version":1,...}
	EnvVar = "DYD_DIAG"
)

type Config struct {
	Version int          `json:"version" yaml:"version"`
	Seed    int64        `json:"seed" yaml:"seed"`
	Rules   []RuleConfig `json:"rules" yaml:"rules"`
}

type RuleConfig struct {
	ID      string       `json:"id,omitempty" yaml:"id,omitempty"`
	Enabled *bool        `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Op      string       `json:"op" yaml:"op"`
	Key     string       `json:"key" yaml:"key"`
	When    WhenConfig   `json:"when" yaml:"when"`
	Action  ActionConfig `json:"action" yaml:"action"`
}

type WhenConfig struct {
	Mode  string `json:"mode" yaml:"mode"`
	Count int64  `json:"count,omitempty" yaml:"count,omitempty"`
	Limit int64  `json:"limit,omitempty" yaml:"limit,omitempty"`
}

type ActionConfig struct {
	Type    string               `json:"type" yaml:"type"`
	Phase   string               `json:"phase,omitempty" yaml:"phase,omitempty"`
	Error   string               `json:"error,omitempty" yaml:"error,omitempty"`
	DelayMS int64                `json:"delay_ms,omitempty" yaml:"delay_ms,omitempty"`
	Output  string               `json:"output,omitempty" yaml:"output,omitempty"` // stdout | stderr | ""
	Capture MetricsCaptureConfig `json:"capture,omitempty" yaml:"capture,omitempty"`
}

type MetricsCaptureConfig struct {
	Calls  *bool `json:"calls,omitempty" yaml:"calls,omitempty"`
	Errors *bool `json:"errors,omitempty" yaml:"errors,omitempty"`
	Timing *bool `json:"timing,omitempty" yaml:"timing,omitempty"`
}

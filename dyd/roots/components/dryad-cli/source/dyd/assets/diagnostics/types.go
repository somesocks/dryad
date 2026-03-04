package diagnostics

const (
	// EnvVar controls diagnostics configuration.
	// Supported values:
	// - file:/path/to/config.yaml
	// - json:{"version":1,...}
	EnvVar = "DYD_DG"
)

type Config struct {
	Version int                 `json:"version" yaml:"version"`
	Seed    int64               `json:"seed" yaml:"seed"`
	Rules   []RuleConfig        `json:"rules" yaml:"rules"`
	Metrics []MetricsRuleConfig `json:"metrics,omitempty" yaml:"metrics,omitempty"`
}

type RuleConfig struct {
	ID      string       `json:"id" yaml:"id"`
	Enabled *bool        `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Op      string       `json:"op" yaml:"op"`
	Key     string       `json:"key" yaml:"key"`
	When    WhenConfig   `json:"when" yaml:"when"`
	Action  ActionConfig `json:"action" yaml:"action"`
	MaxHits int64        `json:"max_hits,omitempty" yaml:"max_hits,omitempty"`
}

type WhenConfig struct {
	Mode    string  `json:"mode" yaml:"mode"`
	Count   int64   `json:"count,omitempty" yaml:"count,omitempty"`
	Percent float64 `json:"percent,omitempty" yaml:"percent,omitempty"`
}

type ActionConfig struct {
	Type    string `json:"type" yaml:"type"`
	Phase   string `json:"phase,omitempty" yaml:"phase,omitempty"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`
	DelayMS int64  `json:"delay_ms,omitempty" yaml:"delay_ms,omitempty"`
}

type MetricsRuleConfig struct {
	ID      string               `json:"id" yaml:"id"`
	Enabled *bool                `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Op      string               `json:"op" yaml:"op"`
	Output  string               `json:"output,omitempty" yaml:"output,omitempty"` // stdout | stderr | ""
	Capture MetricsCaptureConfig `json:"capture,omitempty" yaml:"capture,omitempty"`
}

type MetricsCaptureConfig struct {
	Calls  *bool `json:"calls,omitempty" yaml:"calls,omitempty"`
	Errors *bool `json:"errors,omitempty" yaml:"errors,omitempty"`
	Timing *bool `json:"timing,omitempty" yaml:"timing,omitempty"`
}

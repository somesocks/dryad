package diagnostics

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/yaml"
)

func parseConfigFromEnv(raw string) (Config, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Config{}, fmt.Errorf("%s is empty", EnvVar)
	}

	switch {
	case strings.HasPrefix(raw, "file:"):
		path := strings.TrimSpace(strings.TrimPrefix(raw, "file:"))
		if path == "" {
			return Config{}, fmt.Errorf("%s file: path is empty", EnvVar)
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("read diagnostics file %q: %w", path, err)
		}

		var cfg Config
		if err := yaml.Unmarshal(body, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse diagnostics file %q: %w", path, err)
		}
		return cfg, nil

	case strings.HasPrefix(raw, "json:"):
		doc := strings.TrimSpace(strings.TrimPrefix(raw, "json:"))
		if doc == "" {
			return Config{}, fmt.Errorf("%s json: payload is empty", EnvVar)
		}

		var cfg Config
		if err := json.Unmarshal([]byte(doc), &cfg); err != nil {
			return Config{}, fmt.Errorf("parse diagnostics json payload: %w", err)
		}
		return cfg, nil

	default:
		return Config{}, fmt.Errorf(
			"%s must use file: or json: prefix",
			EnvVar,
		)
	}
}

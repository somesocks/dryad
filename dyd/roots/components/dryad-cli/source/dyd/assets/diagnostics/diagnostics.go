package diagnostics

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
)

var activeEngine atomic.Pointer[engine]
var engineGeneration atomic.Uint64

func Disable() {
	activeEngine.Store(nil)
}

func SetupFromConfig(config Config) error {
	compiled, err := compileConfig(config)
	if err != nil {
		return err
	}
	compiled.version = engineGeneration.Add(1)
	activeEngine.Store(compiled)
	return nil
}

func SetupFromEnv() error {
	raw := strings.TrimSpace(os.Getenv(EnvVar))
	if raw == "" {
		Disable()
		return nil
	}

	config, err := parseConfigFromEnv(raw)
	if err != nil {
		return err
	}

	if err := SetupFromConfig(config); err != nil {
		return fmt.Errorf("compile diagnostics config: %w", err)
	}

	return nil
}

func Apply(point string, key string) error {
	current := activeEngine.Load()
	if current == nil {
		return nil
	}
	r := current.Runner(point)
	if r == nil {
		return nil
	}
	return r(key)
}

package diagnostics

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

var activeEngine atomic.Pointer[engine]
var engineGeneration atomic.Uint64

func Disable() {
	activeEngine.Store(nil)
}

func Reset() {
	Disable()
	ResetMetrics()
	engineGeneration.Store(0)
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
	var result error

	current := activeEngine.Load()
	var metric *compiledMetricsRule
	if current != nil {
		metric = current.Metric(point)
	}
	start := beginMetricsObservation(metric)

	if current == nil {
		result = nil
	} else {
		rules := current.Rules(point)
		if len(rules) == 0 {
			result = nil
		} else {
			result = applyNoopRules(rules, key, 0)
		}
	}

	endMetricsObservation(metric, point, start, result)
	return result
}

func applyNoopRules(rules []*compiledRule, key string, index int) error {
	if index >= len(rules) {
		return nil
	}

	rule := rules[index]

	switch rule.action {
	case actionDelay:
		if rule.matches(key) && rule.delay > 0 {
			time.Sleep(rule.delay)
		}
		return applyNoopRules(rules, key, index+1)

	case actionError:
		hit := rule.matches(key)
		if hit && !rule.postError {
			return rule.err
		}

		err := applyNoopRules(rules, key, index+1)
		if err != nil {
			return err
		}
		if hit && rule.postError {
			return rule.err
		}
		return nil

	default:
		return applyNoopRules(rules, key, index+1)
	}
}

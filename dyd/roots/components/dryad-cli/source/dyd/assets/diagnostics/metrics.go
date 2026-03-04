package diagnostics

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type pointStats struct {
	calls      atomic.Uint64
	errors     atomic.Uint64
	samples    atomic.Uint64
	totalNanos atomic.Uint64
	minNanos   atomic.Uint64
	maxNanos   atomic.Uint64
}

type PointStatsSnapshot struct {
	Calls      uint64 `json:"calls"`
	Errors     uint64 `json:"errors"`
	TotalNanos uint64 `json:"total_nanos"`
	MinNanos   uint64 `json:"min_nanos"`
	MaxNanos   uint64 `json:"max_nanos"`
	AvgNanos   uint64 `json:"avg_nanos"`
}

var metricsRegistry = struct {
	mu     sync.RWMutex
	points map[string]*pointStats
}{
	points: map[string]*pointStats{},
}

func pointStatsFor(point string) *pointStats {
	metricsRegistry.mu.RLock()
	stats := metricsRegistry.points[point]
	metricsRegistry.mu.RUnlock()
	if stats != nil {
		return stats
	}

	metricsRegistry.mu.Lock()
	defer metricsRegistry.mu.Unlock()
	stats = metricsRegistry.points[point]
	if stats != nil {
		return stats
	}

	stats = &pointStats{}
	stats.minNanos.Store(math.MaxUint64)
	metricsRegistry.points[point] = stats
	return stats
}

func observePointInvocation(rule *compiledMetricsRule, point string, elapsed time.Duration, err error) {
	if rule == nil {
		return
	}
	if !rule.captureCalls && !rule.captureErrors && !rule.captureTiming {
		return
	}

	stats := rule.stats.Load()
	if stats == nil {
		stats = pointStatsFor(point)
		if !rule.stats.CompareAndSwap(nil, stats) {
			stats = rule.stats.Load()
		}
	}

	if rule.captureCalls {
		stats.calls.Add(1)
	}
	if rule.captureErrors && err != nil {
		stats.errors.Add(1)
	}
	if !rule.captureTiming {
		return
	}

	nanos := uint64(elapsed)
	stats.samples.Add(1)
	stats.totalNanos.Add(nanos)

	for {
		current := stats.minNanos.Load()
		if nanos >= current {
			break
		}
		if stats.minNanos.CompareAndSwap(current, nanos) {
			break
		}
	}

	for {
		current := stats.maxNanos.Load()
		if nanos <= current {
			break
		}
		if stats.maxNanos.CompareAndSwap(current, nanos) {
			break
		}
	}
}

func MetricsSnapshot() map[string]PointStatsSnapshot {
	metricsRegistry.mu.RLock()
	defer metricsRegistry.mu.RUnlock()

	out := make(map[string]PointStatsSnapshot, len(metricsRegistry.points))
	for point, stats := range metricsRegistry.points {
		calls := stats.calls.Load()
		errors := stats.errors.Load()
		total := stats.totalNanos.Load()
		min := stats.minNanos.Load()
		max := stats.maxNanos.Load()
		samples := stats.samples.Load()
		if samples == 0 || min == math.MaxUint64 {
			min = 0
		}

		avg := uint64(0)
		if samples > 0 {
			avg = total / samples
		}

		out[point] = PointStatsSnapshot{
			Calls:      calls,
			Errors:     errors,
			TotalNanos: total,
			MinNanos:   min,
			MaxNanos:   max,
			AvgNanos:   avg,
		}
	}

	return out
}

func ResetMetrics() {
	metricsRegistry.mu.Lock()
	metricsRegistry.points = map[string]*pointStats{}
	metricsRegistry.mu.Unlock()

	// Keep active compiled metrics rules valid after reset so wrappers can
	// continue writing into fresh point stats without per-call map lookups.
	current := activeEngine.Load()
	if current == nil {
		return
	}
	for point, metric := range current.metrics {
		metric.stats.Store(pointStatsFor(point))
	}
}

func beginMetricsObservation(rule *compiledMetricsRule) (time.Time, bool) {
	if rule == nil {
		return time.Time{}, false
	}
	if !metricsRuleShouldSample(rule) {
		return time.Time{}, false
	}
	if rule.captureTiming {
		return time.Now(), true
	}
	return time.Time{}, true
}

func endMetricsObservation(rule *compiledMetricsRule, point string, sampled bool, start time.Time, err error) {
	if rule == nil {
		return
	}
	if !sampled {
		return
	}
	elapsed := time.Duration(0)
	if rule.captureTiming {
		elapsed = time.Since(start)
	}
	observePointInvocation(rule, point, elapsed, err)
}

func metricsRuleShouldSample(rule *compiledMetricsRule) bool {
	if rule.sampleEvery <= 1 {
		return true
	}
	n := rule.sampleCounter.Add(1)
	return (n & rule.sampleMask) == 0
}

type metricsPointOutput struct {
	Point      string `json:"point"`
	Calls      uint64 `json:"calls"`
	Errors     uint64 `json:"errors"`
	TotalNanos uint64 `json:"total_nanos"`
	MinNanos   uint64 `json:"min_nanos"`
	MaxNanos   uint64 `json:"max_nanos"`
	AvgNanos   uint64 `json:"avg_nanos"`
}

func EmitMetricsOnExit() error {
	current := activeEngine.Load()
	if current == nil {
		return nil
	}
	if len(current.metrics) == 0 {
		return nil
	}

	snapshot := MetricsSnapshot()
	if len(snapshot) == 0 {
		return nil
	}

	points := make([]string, 0, len(current.metrics))
	for point := range current.metrics {
		points = append(points, point)
	}
	sort.Strings(points)

	for _, point := range points {
		stats, ok := snapshot[point]
		if !ok {
			continue
		}

		payload := metricsPointOutput{
			Point:      point,
			Calls:      stats.Calls,
			Errors:     stats.Errors,
			TotalNanos: stats.TotalNanos,
			MinNanos:   stats.MinNanos,
			MaxNanos:   stats.MaxNanos,
			AvgNanos:   stats.AvgNanos,
		}

		line, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		switch current.metrics[point].output {
		case metricsOutputStdout:
			if _, err := fmt.Fprintln(os.Stdout, string(line)); err != nil {
				return err
			}
		default:
			if _, err := fmt.Fprintln(os.Stderr, string(line)); err != nil {
				return err
			}
		}
	}

	return nil
}

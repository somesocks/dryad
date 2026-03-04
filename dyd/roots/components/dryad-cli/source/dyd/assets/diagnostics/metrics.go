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

func observePointInvocation(point string, elapsed time.Duration, err error) {
	current := activeEngine.Load()
	if current == nil {
		return
	}
	if current.Metric(point) == nil {
		return
	}

	stats := pointStatsFor(point)
	nanos := uint64(elapsed)

	stats.calls.Add(1)
	if err != nil {
		stats.errors.Add(1)
	}
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
		if calls == 0 || min == math.MaxUint64 {
			min = 0
		}

		avg := uint64(0)
		if calls > 0 {
			avg = total / calls
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
	defer metricsRegistry.mu.Unlock()
	metricsRegistry.points = map[string]*pointStats{}
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
		if !ok || stats.Calls == 0 {
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

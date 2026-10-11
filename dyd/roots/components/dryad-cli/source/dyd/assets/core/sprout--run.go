package core

import (
	"context"
	"dryad/internal/filepath"
	"dryad/task"
	"fmt"
	"sync"
	"sync/atomic"

	zlog "github.com/rs/zerolog/log"
)

type SproutRunRequest struct {
	WorkingPath       string
	MainOverride      string
	VariantDescriptor string
	Context           string
	Env               map[string]string
	Args              []string
	Stdin             InputConfig
	Stdout            OutputConfig
	Stderr            OutputConfig
	InheritEnv        bool
}

type SproutRunTarget struct {
	Sprout  *SafeSproutReference
	Request SproutRunRequest
}

type SproutsRunRequest struct {
	Targets []SproutRunTarget
	// Stdin applies to the whole invocation, overriding individual target inputs.
	Stdin        InputConfig
	FromStdin    bool
	IgnoreErrors bool
}

type sproutRunExecution struct {
	sprout     *SafeSproutReference
	descriptor string
	request    StemRunRequest
}

func sproutRunLogVariantLabel(descriptor string) string {
	if descriptor == "" {
		return "default"
	}
	return descriptor
}

func (sprout *SafeSproutReference) prepareRun(req SproutRunRequest) ([]sproutRunExecution, error) {
	if err := req.Stdout.validate(); err != nil {
		return nil, fmt.Errorf("stdout: %w", err)
	}
	if err := req.Stderr.validate(); err != nil {
		return nil, fmt.Errorf("stderr: %w", err)
	}
	err, selector := resolveSproutRunVariantSelector(req.VariantDescriptor)
	if err != nil {
		return nil, err
	}
	err, available := sprout.runStemVariants()
	if err != nil {
		return nil, err
	}
	err, selected := resolveSproutRunStemVariants(available, selector)
	if err != nil {
		return nil, err
	}
	relPath, err := filepath.Rel(sprout.Sprouts.Garden.BasePath, sprout.BasePath)
	if err != nil {
		return nil, err
	}
	executions := make([]sproutRunExecution, 0, len(selected))
	for _, variant := range selected {
		stdout, stderr := req.Stdout, req.Stderr
		logBase := "dyd-sprout-run--" + sanitizePathSegment(relPath)
		if len(selected) > 1 || variant.DescriptorRaw != "" {
			logBase += "--" + sanitizePathSegment(sproutRunLogVariantLabel(variant.DescriptorRaw))
		}
		if stdout.Mode == OutputLog && stdout.Name == "" {
			stdout.Name = logBase + ".out"
		}
		if stderr.Mode == OutputLog && stderr.Name == "" {
			stderr.Name = logBase + ".err"
		}
		executions = append(executions, sproutRunExecution{
			sprout:     sprout,
			descriptor: variant.DescriptorRaw,
			request: StemRunRequest{
				Garden:       sprout.Sprouts.Garden,
				StemPath:     variant.StemPath,
				WorkingPath:  req.WorkingPath,
				MainOverride: req.MainOverride,
				Context:      req.Context,
				Env:          req.Env,
				Args:         req.Args,
				Stdin:        req.Stdin,
				Stdout:       stdout,
				Stderr:       stderr,
				InheritEnv:   req.InheritEnv,
			},
		})
	}
	return executions, nil
}

func (execution sproutRunExecution) run() error {
	label := sproutRunLogVariantLabel(execution.descriptor)
	zlog.Info().Str("sprout", execution.sprout.BasePath).Str("variant", label).Msg("sprout run starting")
	if err := StemRun(execution.request); err != nil {
		if execution.descriptor != "" {
			return fmt.Errorf("error running sprout variant %s: %w", execution.descriptor, err)
		}
		return err
	}
	zlog.Info().Str("sprout", execution.sprout.BasePath).Str("variant", label).Msg("sprout run finished")
	return nil
}

// SproutsRun prepares every selected variant before resolving stdin ownership.
// This prevents separate sprouts from each treating themselves as a single run.
func SproutsRun(ctx context.Context, req SproutsRunRequest) error {
	parallel, err := task.Concurrency(ctx)
	if err != nil {
		return err
	}
	if _, err := req.Stdin.Resolve(0, parallel, req.FromStdin); err != nil {
		return err
	}
	type preparedTarget struct {
		target     SproutRunTarget
		executions []sproutRunExecution
		started    sync.Once
		failed     atomic.Bool
		remaining  atomic.Int64
	}
	prepared := make([]*preparedTarget, 0, len(req.Targets))
	count := 0
	for _, target := range req.Targets {
		executions, err := target.Sprout.prepareRun(target.Request)
		if err != nil {
			zlog.Warn().Str("sprout", target.Sprout.BasePath).Err(err).Msg("error preparing sprout run")
			if !req.IgnoreErrors {
				return err
			}
			continue
		}
		count += len(executions)
		group := &preparedTarget{target: target, executions: executions}
		group.remaining.Store(int64(len(executions)))
		prepared = append(prepared, group)
	}
	stdin, err := req.Stdin.Resolve(count, parallel, req.FromStdin)
	if err != nil {
		return err
	}
	type scheduledExecution struct {
		group     *preparedTarget
		execution sproutRunExecution
	}
	scheduled := make([]scheduledExecution, 0, count)
	for _, group := range prepared {
		for _, execution := range group.executions {
			execution.request.Stdin = stdin
			scheduled = append(scheduled, scheduledExecution{group: group, execution: execution})
		}
	}
	runExecution := func(ctx context.Context, job scheduledExecution) (error, any) {
		group := job.group
		// IgnoreErrors applies to sprouts: pending variants of a failed target
		// are skipped, while other targets may continue. Started work is awaited.
		if group.failed.Load() {
			return nil, nil
		}
		label := sproutRunLogVariantLabel(group.target.Request.VariantDescriptor)
		path := group.target.Sprout.BasePath
		group.started.Do(func() {
			zlog.Info().Str("sprout", path).Str("variant_selector", label).Msg("sprout run requested")
		})
		if err := job.execution.run(); err != nil {
			group.failed.Store(true)
			zlog.Warn().Str("sprout", path).Str("variant_selector", label).Err(err).Msg("sprout threw error during execution")
			if !req.IgnoreErrors {
				return err, nil
			}
			return nil, nil
		}
		if group.remaining.Add(-1) == 0 {
			zlog.Info().Str("sprout", path).Str("variant_selector", label).Msg("sprout run completed")
		}
		return nil, nil
	}
	// Schedule only concrete executions. Nested sprout tasks would occupy slots
	// while waiting for their variants, reducing usable process concurrency.
	err, _ = task.ParallelMap(runExecution)(ctx, scheduled)
	return err
}

func (sprout *SafeSproutReference) Run(ctx context.Context, req SproutRunRequest) error {
	return SproutsRun(ctx, SproutsRunRequest{
		Targets: []SproutRunTarget{{Sprout: sprout, Request: req}},
		Stdin:   req.Stdin,
	})
}

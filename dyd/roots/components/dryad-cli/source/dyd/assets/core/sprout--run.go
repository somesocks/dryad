package core

import (
	"context"
	"dryad/internal/filepath"
	"dryad/task"
	"fmt"

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
	}
	prepared := make([]preparedTarget, 0, len(req.Targets))
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
		prepared = append(prepared, preparedTarget{target: target, executions: executions})
	}
	stdin, err := req.Stdin.Resolve(count, parallel, req.FromStdin)
	if err != nil {
		return err
	}
	runTarget := func(ctx context.Context, target preparedTarget) (error, any) {
		label := sproutRunLogVariantLabel(target.target.Request.VariantDescriptor)
		path := target.target.Sprout.BasePath
		zlog.Info().Str("sprout", path).Str("variant_selector", label).Msg("sprout run requested")
		for _, execution := range target.executions {
			execution.request.Stdin = stdin
			if err := execution.run(); err != nil {
				zlog.Warn().Str("sprout", path).Str("variant_selector", label).Err(err).Msg("sprout threw error during execution")
				if !req.IgnoreErrors {
					return err, nil
				}
				return nil, nil
			}
		}
		zlog.Info().Str("sprout", path).Str("variant_selector", label).Msg("sprout run completed")
		return nil, nil
	}
	// Retain the existing scheduling until the concurrency change: garden targets
	// can overlap, but variants and stdin-supplied targets are still serial.
	if req.FromStdin {
		for _, target := range prepared {
			if err, _ := runTarget(ctx, target); err != nil {
				return err
			}
		}
		return nil
	}
	err, _ = task.ParallelMap(runTarget)(ctx, prepared)
	return err
}

func (sprout *SafeSproutReference) Run(ctx context.Context, req SproutRunRequest) error {
	return SproutsRun(ctx, SproutsRunRequest{
		Targets: []SproutRunTarget{{Sprout: sprout, Request: req}},
		Stdin:   req.Stdin,
	})
}

package core

import (
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
	JoinStdout        bool
	LogStdout         struct {
		Path string
		Name string
	}
	JoinStderr bool
	LogStderr  struct {
		Path string
		Name string
	}
	InheritEnv bool
}

func sproutRunLogVariantLabel(descriptor string) string {
	if descriptor == "" {
		return "default"
	}

	return descriptor
}

func (sprout *SafeSproutReference) Run(
	ctx *task.ExecutionContext,
	req SproutRunRequest,
) error {
	err, variantSelector := resolveSproutRunVariantSelector(req.VariantDescriptor)
	if err != nil {
		return err
	}

	err, availableVariants := sprout.runStemVariants()
	if err != nil {
		return err
	}

	err, selectedVariants := resolveSproutRunStemVariants(availableVariants, variantSelector)
	if err != nil {
		return err
	}

	relSproutPath, err := filepath.Rel(
		sprout.Sprouts.Garden.BasePath,
		sprout.BasePath,
	)
	if err != nil {
		return err
	}

	for _, selectedVariant := range selectedVariants {
		logStdout := req.LogStdout
		logStderr := req.LogStderr
		variantLabel := sproutRunLogVariantLabel(selectedVariant.DescriptorRaw)

		if logStdout.Path != "" && logStdout.Name == "" {
			logStdout.Name = "dyd-sprout-run--" + sanitizePathSegment(relSproutPath)
			if len(selectedVariants) > 1 || selectedVariant.DescriptorRaw != "" {
				suffix := selectedVariant.DescriptorRaw
				if suffix == "" {
					suffix = "default"
				}
				logStdout.Name = logStdout.Name + "--" + sanitizePathSegment(suffix)
			}
			logStdout.Name = logStdout.Name + ".out"
		}

		if logStderr.Path != "" && logStderr.Name == "" {
			logStderr.Name = "dyd-sprout-run--" + sanitizePathSegment(relSproutPath)
			if len(selectedVariants) > 1 || selectedVariant.DescriptorRaw != "" {
				suffix := selectedVariant.DescriptorRaw
				if suffix == "" {
					suffix = "default"
				}
				logStderr.Name = logStderr.Name + "--" + sanitizePathSegment(suffix)
			}
			logStderr.Name = logStderr.Name + ".err"
		}

		zlog.Info().
			Str("sprout", sprout.BasePath).
			Str("variant", variantLabel).
			Msg("sprout run starting")

		err = StemRun(
			StemRunRequest{
				Garden:       sprout.Sprouts.Garden,
				StemPath:     selectedVariant.StemPath,
				WorkingPath:  req.WorkingPath,
				MainOverride: req.MainOverride,
				Context:      req.Context,
				Env:          req.Env,
				Args:         req.Args,
				JoinStdout:   req.JoinStdout,
				LogStdout:    logStdout,
				JoinStderr:   req.JoinStderr,
				LogStderr:    logStderr,
				InheritEnv:   req.InheritEnv,
			},
		)
		if err != nil {
			if selectedVariant.DescriptorRaw != "" {
				return fmt.Errorf("error running sprout variant %s: %w", selectedVariant.DescriptorRaw, err)
			}
			return err
		}

		zlog.Info().
			Str("sprout", sprout.BasePath).
			Str("variant", variantLabel).
			Msg("sprout run finished")
	}

	return err
}

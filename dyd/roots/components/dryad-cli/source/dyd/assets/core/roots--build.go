package core

import (
	// "dryad/internal/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

type rootsBuildRequest struct {
	Roots             *SafeRootsReference
	Filter            RootVariantFilter
	VariantDescriptor string
	JoinStdout        bool
	JoinStderr        bool
	LogStdout         struct {
		Path string
		Name string
	}
	LogStderr struct {
		Path string
		Name string
	}
}

func rootsBuild(ctx *task.ExecutionContext, request rootsBuildRequest) (error, any) {
	var err error
	var sprouts *SafeSproutsReference

	zlog.Debug().
		Str("gardenPath", request.Roots.Garden.BasePath).
		Msg("RootsBuild")

	err, sprouts = request.Roots.Garden.Sprouts().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	// prune sprouts before build
	err = sprouts.Prune(ctx)
	if err != nil {
		return err, nil
	}

	var buildRoot = func(ctx *task.ExecutionContext, root *SafeRootReference) (error, any) {
		err, variantSelectorDescriptor := normalizeRootBuildVariantDescriptor(request.VariantDescriptor)
		if err != nil {
			return err, nil
		}

		err, variantSelector := variantDescriptorParseFilesystem(variantSelectorDescriptor)
		if err != nil {
			return err, nil
		}

		err, variants := root.ResolveBuildVariantReferences(
			ctx,
			RootResolveBuildVariantsRequest{
				Selector:                variantSelector,
				IgnoreUnknownDimensions: true,
			},
		)
		if err != nil {
			return err, nil
		}

		selectedVariants := make([]*SafeRootVariantReference, 0, len(variants))
		for _, variant := range variants {
			err, shouldMatch := request.Filter(ctx, variant)
			if err != nil {
				return err, nil
			}
			if !shouldMatch {
				continue
			}

			selectedVariants = append(selectedVariants, variant)
		}

		if len(selectedVariants) == 0 {
			return nil, nil
		}

		err, _ = root.buildSproutResolvedVariants(
			ctx,
			rootBuildSproutResolvedVariantsRequest{
				Variants:   selectedVariants,
				JoinStdout: request.JoinStdout,
				JoinStderr: request.JoinStderr,
				LogStdout:  request.LogStdout,
				LogStderr:  request.LogStderr,
			},
		)
		return err, nil
	}

	// build each root in the garden
	err = request.Roots.Walk(
		ctx,
		RootsWalkRequest{
			OnMatch: buildRoot,
		},
	)

	return err, nil
}

type RootsBuildRequest struct {
	Filter            RootVariantFilter
	VariantDescriptor string
	JoinStdout        bool
	JoinStderr        bool
	LogStdout         struct {
		Path string
		Name string
	}
	LogStderr struct {
		Path string
		Name string
	}
}

func (roots *SafeRootsReference) Build(ctx *task.ExecutionContext, req RootsBuildRequest) error {
	err, _ := rootsBuild(
		ctx,
		rootsBuildRequest{
			Roots:             roots,
			Filter:            req.Filter,
			VariantDescriptor: req.VariantDescriptor,
			JoinStdout:        req.JoinStdout,
			JoinStderr:        req.JoinStderr,
			LogStdout:         req.LogStdout,
			LogStderr:         req.LogStderr,
		},
	)

	return err
}

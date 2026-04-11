package core

import (
	"dryad/task"
	// zlog "github.com/rs/zerolog/log"
)

type RootReplaceRequest struct {
	Filter RootVariantFilter
	Source RootReplaceTargetSpec
	Dest   RootReplaceTargetSpec
}

func (root *SafeRootReference) Replace(ctx *task.ExecutionContext, request RootReplaceRequest) error {
	var err error
	seenRequirements := map[string]struct{}{}

	var onRootRequirement = func(ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any) {
		if _, seen := seenRequirements[requirement.BasePath]; seen {
			return nil, nil
		}
		seenRequirements[requirement.BasePath] = struct{}{}

		var targetSpec *RootRequirementTargetSpec
		var err error

		err, targetSpec = requirement.TargetSpec(ctx)
		if err != nil {
			return err, nil
		}

		if rootRequirementTargetSpecMatchesReplaceTarget(targetSpec, request.Source) {
			err, replacedTargetSpec := rootRequirementTargetSpecApplyReplaceTarget(
				targetSpec,
				request.Dest,
			)
			if err != nil {
				return err, nil
			}

			err = requirement.Replace(ctx, replacedTargetSpec)
			if err != nil {
				return err, nil
			}
		}

		return nil, nil
	}

	var onVariant = func(ctx *task.ExecutionContext, variant *SafeRootVariantReference) (error, any) {
		if variant.Requirements == nil {
			return nil, nil
		}

		err := variant.Requirements.Walk(
			ctx,
			RootRequirementsWalkRequest{
				OnMatch: onRootRequirement,
			},
		)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	err = root.Roots.WalkVariants(
		ctx,
		RootsWalkVariantsRequest{
			ShouldMatch: request.Filter,
			OnMatch:     onVariant,
		},
	)
	return err
}

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
	filter := request.Filter
	if filter == nil {
		filter = func(*task.ExecutionContext, *SafeRootVariantReference) (error, bool) {
			return nil, true
		}
	}

	var onRootRequirement = func(ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any) {
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

	var onRoot = func(ctx *task.ExecutionContext, candidateRoot *SafeRootReference) (error, any) {
		err, variants := candidateRoot.ResolveBuildVariantReferences(
			ctx,
			RootResolveBuildVariantsRequest{},
		)
		if err != nil {
			return err, nil
		}

		matchedVariants := make([]*SafeRootVariantReference, 0, len(variants))
		for _, variant := range variants {
			err, shouldMatch := filter(ctx, variant)
			if err != nil {
				return err, nil
			}
			if shouldMatch {
				matchedVariants = append(matchedVariants, variant)
			}
		}

		if len(matchedVariants) == 0 {
			return nil, nil
		}

		selectedRequirementsPaths := map[string]struct{}{}
		for _, variant := range matchedVariants {
			if variant.Requirements == nil {
				continue
			}
			selectedRequirementsPaths[variant.Requirements.BasePath] = struct{}{}
		}

		err = candidateRoot.WalkAllRequirements(
			ctx,
			RootWalkAllRequirementsRequest{
				OnMatch: func(ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any) {
					if _, exists := selectedRequirementsPaths[requirement.Requirements.BasePath]; !exists {
						return nil, nil
					}

					return onRootRequirement(ctx, requirement)
				},
			},
		)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	err := root.Roots.Walk(
		ctx,
		RootsWalkRequest{
			OnMatch: onRoot,
		},
	)
	return err
}

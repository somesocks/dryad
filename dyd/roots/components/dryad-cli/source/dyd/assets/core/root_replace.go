package core

import (
	"dryad/task"

	// zlog "github.com/rs/zerolog/log"
)

type RootReplaceRequest struct {
	Dest *SafeRootReference
}

func (root *SafeRootReference) Replace(ctx *task.ExecutionContext, request RootReplaceRequest) (error) {
	var err error

	var onRootRequirement = func (ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any) {
		var target *SafeRootReference
		var err error

		err, target = requirement.Target(ctx)
		if err != nil {
			return err, nil
		}

		if target.BasePath == root.BasePath {
			err = requirement.Replace(ctx, request.Dest)
			if err != nil {
				return err, nil
			}	
		}

		return nil, nil
	}

	var onRoot = func (ctx *task.ExecutionContext, root *SafeRootReference) (error, any) {
		var requirements *SafeRootRequirementsReference
		var err error

		err, requirements = root.Requirements().Resolve(ctx)
		if err != nil {
			return err, nil
		} else if requirements == nil {
			// do nothing if there are no requirements
			return nil, nil
		}

		err = requirements.Walk(
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

	err = root.Roots.Walk(
		ctx,
		RootsWalkRequest{
			OnMatch: onRoot,
		},
	)
	return err
}
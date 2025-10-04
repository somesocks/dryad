package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"
	// "path/filepath"
)

type RootRequirementsWalkRequest struct {
	OnMatch func (ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any)
}

func (requirements *SafeRootRequirementsReference) Walk(ctx *task.ExecutionContext, req RootRequirementsWalkRequest) error {
	var err error

	var shouldWalk = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		return nil, node.Path == node.BasePath
	}

	var shouldMatch = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		return nil, node.Path != node.BasePath
	}

	var onMatch = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		var unsafeRequirementRef = UnsafeRootRequirementReference{
			BasePath: node.Path,
			Requirements: requirements,
		}
		var safeRequirementRef *SafeRootRequirementReference
		var err error

		err, safeRequirementRef = unsafeRequirementRef.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, _ = req.OnMatch(ctx, safeRequirementRef)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	onMatch = dydfs.ConditionalWalkAction(onMatch, shouldMatch)

	err, _ = dydfs.Walk6(
		ctx,
		dydfs.Walk6Request{
			BasePath: requirements.BasePath,
			Path:     requirements.BasePath,
			VPath:    requirements.BasePath,
			ShouldWalk: shouldWalk,
			OnPreMatch: onMatch,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

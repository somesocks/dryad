package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	// "path/filepath"
)

type RootRequirementsWalkRequest struct {
	OnMatch func (ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any)
}

func (requirements *SafeRootRequirementsReference) Walk(ctx *task.ExecutionContext, req RootRequirementsWalkRequest) error {
	var err error

	var shouldCrawl = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		return nil, node.Path == node.BasePath
	}

	var shouldMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		return nil, node.Path != node.BasePath
	}

	var onMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
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

	err, _ = fs2.BFSWalk3(
		ctx,
		fs2.Walk5Request{
			Path:     requirements.BasePath,
			VPath:    requirements.BasePath,
			BasePath: requirements.BasePath,
			ShouldCrawl: shouldCrawl,
			ShouldMatch: shouldMatch,
			OnMatch: onMatch,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

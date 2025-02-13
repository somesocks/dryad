package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	"path/filepath"
)

type RootRequirementsWalkRequest struct {
	Root *SafeRootReference
	OnMatch func (ctx *task.ExecutionContext, requirement *SafeRootReference) (error, any)
}

func RootRequirementsWalk(
	ctx *task.ExecutionContext,
	req RootRequirementsWalkRequest,
) (error, any) {
	var requirementsPath string = filepath.Join(req.Root.BasePath, "dyd", "requirements")
	var requirementsExists bool
	var err error

	requirementsExists, err = fileExists(requirementsPath)
	if err != nil {
		return err, nil
	}

	// if requirements doesn't exist, do nothing
	if !requirementsExists {
		return nil, nil
	}

	var shouldCrawl = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		return nil, node.Path == node.BasePath
	}

	var shouldMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		return nil, node.Path != node.BasePath
	}

	var onMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
		var unsafeRequirementRef = UnsafeRootReference{
			BasePath: node.Path,
			Garden: req.Root.Garden,
		}
		var safeRequirementRef SafeRootReference
		var err error

		err, safeRequirementRef = unsafeRequirementRef.Resolve(ctx, nil)
		if err != nil {
			return err, nil
		}

		err, _ = req.OnMatch(ctx, &safeRequirementRef)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	err, _ = fs2.BFSWalk3(
		ctx,
		fs2.Walk5Request{
			Path:     requirementsPath,
			VPath:    requirementsPath,
			BasePath: requirementsPath,
			ShouldCrawl: shouldCrawl,
			ShouldMatch: shouldMatch,
			OnMatch: onMatch,
		},
	)
	if err != nil {
		return err, nil
	}

	return nil, nil
} 

package core

import (
	"context"
	dydfs "dryad/filesystem"
	// "dryad/internal/filepath"
)

type RootRequirementsWalkRequest struct {
	OnMatch func(ctx context.Context, requirement *SafeRootRequirementReference) (error, any)
}

func (requirements *SafeRootRequirementsReference) Walk(ctx context.Context, req RootRequirementsWalkRequest) error {
	var err error

	var shouldWalk = func(ctx context.Context, node dydfs.Walk6Node) (error, bool) {
		return nil, node.Path == node.BasePath
	}

	var shouldMatch = func(ctx context.Context, node dydfs.Walk6Node) (error, bool) {
		return nil, node.Path != node.BasePath
	}

	var onMatch = func(ctx context.Context, node dydfs.Walk6Node) (error, any) {
		safeRequirementRef := &SafeRootRequirementReference{
			BasePath:     node.Path,
			fileInfo:     node.Info,
			Requirements: requirements,
		}

		err, _ := req.OnMatch(ctx, safeRequirementRef)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	onMatch = dydfs.ConditionalWalkAction(onMatch, shouldMatch)

	err, _ = dydfs.Walk6(
		ctx,
		dydfs.Walk6Request{
			BasePath:   requirements.BasePath,
			Path:       requirements.BasePath,
			VPath:      requirements.BasePath,
			ShouldWalk: shouldWalk,
			OnPreMatch: onMatch,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

package core

import (
	// "path/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"

)

type rootsBuildRequest struct {
	Roots *SafeRootsReference
	Filter func (*task.ExecutionContext, *SafeRootReference) (error, bool)
	JoinStdout bool
	JoinStderr bool
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

	var buildRoot = func (ctx *task.ExecutionContext, root *SafeRootReference) (error, any) {

		var err error
		var shouldMatch bool

		err, shouldMatch = request.Filter(ctx, root)
		if err != nil {
			return err, nil
		}

		// if the root isn't being excluded by a selector, build it
		if shouldMatch {
			err, _ = root.Build(
				ctx,
				RootBuildRequest{
					JoinStdout: request.JoinStdout,
					JoinStderr: request.JoinStderr,
				},
			)
			return err, nil
		} else {
			return nil, nil
		}
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
	Filter func (*task.ExecutionContext, *SafeRootReference) (error, bool)
	JoinStdout bool
	JoinStderr bool
}

func (roots *SafeRootsReference) Build(ctx *task.ExecutionContext, req RootsBuildRequest) (error) {
	err, _ := rootsBuild(
		ctx,
		rootsBuildRequest{
			Roots: roots,
			Filter: req.Filter,
			JoinStdout: req.JoinStdout,
			JoinStderr: req.JoinStderr,
		},
	)

	return err
}
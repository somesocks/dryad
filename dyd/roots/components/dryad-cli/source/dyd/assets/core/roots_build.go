package core

import (
	"path/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"

)

type RootsBuildRequest struct {
	Garden *SafeGardenReference
	IncludeRoots func(string) bool
	ExcludeRoots func(string) bool
	JoinStdout bool
	JoinStderr bool
}

func RootsBuild(ctx *task.ExecutionContext, request RootsBuildRequest) (error, any) {
	var err error

	zlog.Debug().
		Str("gardenPath", request.Garden.BasePath).
		Msg("RootsBuild")

	// prune sprouts before build
	err = SproutsPrune(request.Garden)
	if err != nil {
		return err, nil
	}

	var buildRoot = func (ctx *task.ExecutionContext, match *SafeRootReference) (error, any) {
		// calculate the relative path to the root from the base of the garden
		relPath, err := filepath.Rel(match.Garden.BasePath, match.BasePath)
		if err != nil {
			return err, nil
		}

		// if the root isn't being excluded by a selector, build it
		if request.IncludeRoots(relPath) && !request.ExcludeRoots(relPath) {
			err, _ = RootBuild(
				ctx,
				RootBuildRequest{
					Garden: request.Garden,
					RootPath: match.BasePath,
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
	err, _ = RootsWalk(
		ctx,
		RootsWalkRequest{
			Garden: request.Garden,
			OnMatch: buildRoot,
		},
	)

	return err, nil
}

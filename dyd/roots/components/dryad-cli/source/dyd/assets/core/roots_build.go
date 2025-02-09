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
	var gardenPath string

	zlog.Debug().
		Str("gardenPath", request.Garden.BasePath).
		Msg("RootsBuild")

	gardenPath = request.Garden.BasePath

	// prune sprouts before build
	err = SproutsPrune(gardenPath)
	if err != nil {
		return err, nil
	}

	var buildRoot = func (ctx *task.ExecutionContext, match RootsWalkMatch) (error, any) {
		// calculate the relative path to the root from the base of the garden
		relPath, err := filepath.Rel(match.GardenPath, match.RootPath)
		if err != nil {
			return err, nil
		}

		// if the root isn't being excluded by a selector, build it
		if request.IncludeRoots(relPath) && !request.ExcludeRoots(relPath) {
			err, _ = RootBuild(
				ctx,
				RootBuildRequest{
					Garden: request.Garden,
					RootPath: match.RootPath,
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
			GardenPath: gardenPath,
			OnRoot: buildRoot,
		},
	)

	return err, nil
}

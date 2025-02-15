package core

import (
	"path/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"

)

type rootsBuildRequest struct {
	Roots *SafeRootsReference
	IncludeRoots func(string) bool
	ExcludeRoots func(string) bool
	JoinStdout bool
	JoinStderr bool
}

func rootsBuild(ctx *task.ExecutionContext, request rootsBuildRequest) (error, any) {
	var err error

	zlog.Debug().
		Str("gardenPath", request.Roots.Garden.BasePath).
		Msg("RootsBuild")

	// prune sprouts before build
	err = SproutsPrune(request.Roots.Garden)
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
					Root: match,
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
	IncludeRoots func(string) bool
	ExcludeRoots func(string) bool
	JoinStdout bool
	JoinStderr bool
}

func (roots *SafeRootsReference) Build(ctx *task.ExecutionContext, req RootsBuildRequest) (error) {
	err, _ := rootsBuild(
		ctx,
		rootsBuildRequest{
			Roots: roots,
			IncludeRoots: req.IncludeRoots,
			ExcludeRoots: req.ExcludeRoots,
			JoinStdout: req.JoinStdout,
			JoinStderr: req.JoinStderr,
		},
	)

	return err
}
package core

import (
	// "path/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"

)

type rootsBuildRequest struct {
	Roots *SafeRootsReference
	IncludeRoots []string
	ExcludeRoots []string
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

		var matchesInclude = false
		var matchesExclude = false

		if len(request.IncludeRoots) == 0 { matchesInclude = true }

		for _, include := range request.IncludeRoots {
			var matchesFilter bool
			err, matchesFilter = root.Filter(
				ctx,
				RootFilterRequest{
					Expression: include,
				},
			)
			if err != nil {
				return err, nil
			}
			matchesInclude = matchesInclude || matchesFilter
			if matchesInclude {
				break
			}
		}

		for _, exclude := range request.ExcludeRoots {
			var matchesFilter bool
			err, matchesFilter = root.Filter(
				ctx,
				RootFilterRequest{
					Expression: exclude,
				},
			)
			if err != nil {
				return err, nil
			}
			matchesExclude = matchesExclude || matchesFilter
			if matchesExclude {
				break
			}
		}

		// if the root isn't being excluded by a selector, build it
		if matchesInclude && !matchesExclude {
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
	IncludeRoots []string
	ExcludeRoots []string
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
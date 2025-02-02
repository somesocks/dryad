package core

import (
	"path/filepath"

	"dryad/task"
)

type RootsBuildRequest struct {
	GardenPath     string
	IncludeRoots func(string) bool
	ExcludeRoots func(string) bool
	JoinStdout bool
	JoinStderr bool
}

func RootsBuild(ctx *task.ExecutionContext, request RootsBuildRequest) (error, any) {
	// fmt.Println("[trace] GardenBuild", request.BasePath)
	var err error

	gardenPath := request.GardenPath

	// handle relative garden paths
	gardenPath, err = filepath.Abs(gardenPath)
	if err != nil {
		return err, nil
	}

	// fmt.Println("[trace] GardenBuild gardenPath 1", gardenPath)

	// make sure it points to the base of the garden path
	gardenPath, err = GardenPath(gardenPath)
	if err != nil {
		return err, nil
	}

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

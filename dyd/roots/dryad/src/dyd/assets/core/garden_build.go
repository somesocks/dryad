package core

import (
	"io/fs"
	"path/filepath"
)

type GardenBuildRequest struct {
	BasePath     string
	IncludeRoots func(string) bool
	ExcludeRoots func(string) bool
}

func GardenBuild(context BuildContext, request GardenBuildRequest) error {
	var err error

	gardenPath := request.BasePath

	// handle relative garden paths
	gardenPath, err = filepath.Abs(gardenPath)
	if err != nil {
		return err
	}

	// make sure it points to the base of the garden path
	gardenPath, err = GardenPath(gardenPath)
	if err != nil {
		return err
	}

	var rootsPath = filepath.Join(gardenPath, "dyd", "roots")

	// build each root in the garden
	err = GardenRootsWalk(
		rootsPath,
		func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// calculate the relative path to the root from the base of the garden
			relPath, err := filepath.Rel(gardenPath, path)
			if err != nil {
				return err
			}

			// if the root isn't being excluded by a selector, build it
			if request.IncludeRoots(relPath) && !request.ExcludeRoots(relPath) {
				_, err = RootBuild(context, path)
				return err
			} else {
				return nil
			}

		},
	)

	return err
}

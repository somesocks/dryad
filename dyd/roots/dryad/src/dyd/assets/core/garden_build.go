package core

import (
	"io/fs"
	"path/filepath"
)

func GardenBuild(context BuildContext, gardenPath string) error {
	var err error

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

			_, err = RootBuild(context, path)
			return err
		},
	)

	return err
}

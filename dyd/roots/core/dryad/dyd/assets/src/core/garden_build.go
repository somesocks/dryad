package core

import (
	"io/fs"
	"path/filepath"
)

func GardenBuild(context BuildContext, gardenPath string) error {

	var err error
	gardenPath, err = GardenPath(gardenPath)
	if err != nil {
		return err
	}

	var rootsPath = filepath.Join(gardenPath, "dyd", "roots")

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

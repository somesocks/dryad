package core

import (
	"path/filepath"
)

func GardenBuild(gardenPath string) error {

	var err error
	gardenPath, err = GardenPath(gardenPath)
	if err != nil {
		return err
	}

	var rootsPath = filepath.Join(gardenPath, "dyd", "roots")

	var aliases []string
	aliases, err = filepath.Glob(rootsPath + "/*")
	if err != nil {
		return err
	}
	//
	// build all dependencies
	for _, alias := range aliases {
		_, err = RootBuild(alias)
		if err != nil {
			return err
		}
	}

	return nil
}

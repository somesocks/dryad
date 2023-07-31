package core

import (
	"path/filepath"
)

func RootsPath(path string) (string, error) {
	var gardenPath, err = GardenPath(path)
	if err != nil {
		return "", err
	}

	return filepath.Join(gardenPath, "dyd", "roots"), nil
}

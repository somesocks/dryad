package core

import (
	"path/filepath"
)

func SproutsPath(path string) (string, error) {
	var gardenPath, err = GardenPath(path)
	if err != nil {
		return "", err
	}

	return filepath.Join(gardenPath, "dyd", "sprouts"), nil
}

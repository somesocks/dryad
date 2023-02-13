package core

import (
	"path/filepath"
)

func ScopesPath(path string) (string, error) {
	gardenPath, err := GardenPath(path)
	if err != nil {
		return "", err
	}

	return filepath.Join(gardenPath, "dyd", "shed", "scopes"), nil
}

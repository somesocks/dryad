package core

import (
	"path/filepath"
)

func HeapPath(path string) (string, error) {
	gardenPath, err := GardenPath(path)

	if err != nil {
		return "", err
	}

	heapPath := filepath.Join(gardenPath, "dyd", "heap")
	return heapPath, nil
}

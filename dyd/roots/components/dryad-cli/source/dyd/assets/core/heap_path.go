package core

import (
	"os"
	"path/filepath"
)

func HeapPath(path string) (string, error) {
	gardenPath, err := GardenPath(path)

	if err != nil {
		return "", err
	}

	heapPath := filepath.Join(gardenPath, "dyd", "heap")
	_, err = os.Stat(heapPath)
	if err != nil {
		err, _ = GardenCreate(GardenCreateRequest{BasePath: gardenPath})
	}

	if err != nil {
		return "", err
	}

	return heapPath, nil
}

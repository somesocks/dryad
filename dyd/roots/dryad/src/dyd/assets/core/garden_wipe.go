package core

import (
	"os"
	"path/filepath"
)

func GardenWipe(gardenPath string) error {
	gardenPath, err := GardenPath(gardenPath)
	if err != nil {
		return err
	}

	sproutsPath := filepath.Join(gardenPath, "dyd", "sprouts")
	heapPath := filepath.Join(gardenPath, "dyd", "heap")

	err = os.RemoveAll(sproutsPath)
	if err != nil {
		return err
	}

	err = os.RemoveAll(heapPath)
	if err != nil {
		return err
	}

	err = GardenInit(gardenPath)
	if err != nil {
		return err
	}

	return nil
}

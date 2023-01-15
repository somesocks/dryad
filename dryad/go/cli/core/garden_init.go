package core

import (
	"os"
	"path/filepath"
)

func GardenInit(path string) error {
	var gardenPath string = filepath.Join(path, "dyd")
	if err := os.MkdirAll(gardenPath, os.ModePerm); err != nil {
		return err
	}

	var heapPath string = filepath.Join(gardenPath, "heap")
	if err := os.MkdirAll(heapPath, os.ModePerm); err != nil {
		return err
	}

	var rootsPath string = filepath.Join(gardenPath, "roots")
	if err := os.MkdirAll(rootsPath, os.ModePerm); err != nil {
		return err
	}

	var sproutsPath string = filepath.Join(gardenPath, "sprouts")
	if err := os.MkdirAll(sproutsPath, os.ModePerm); err != nil {
		return err
	}

	var stemsPath string = filepath.Join(gardenPath, "garden")
	if err := os.MkdirAll(stemsPath, os.ModePerm); err != nil {
		return err
	}

	derivationsPath := filepath.Join(gardenPath, "derivations")
	if err := os.MkdirAll(derivationsPath, os.ModePerm); err != nil {
		return err
	}

	return nil
}

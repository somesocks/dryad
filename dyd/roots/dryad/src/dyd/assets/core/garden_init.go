package core

import (
	"os"
	"path/filepath"
)

func GardenInit(path string) error {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}

	gardenPath := filepath.Join(path, "dyd")
	if err := os.MkdirAll(gardenPath, os.ModePerm); err != nil {
		return err
	}

	flagPath := filepath.Join(gardenPath, "garden")
	if _, err := os.Create(flagPath); err != nil {
		return err
	}

	heapPath := filepath.Join(gardenPath, "heap")
	if err := os.MkdirAll(heapPath, os.ModePerm); err != nil {
		return err
	}

	heapFilesPath := filepath.Join(heapPath, "files")
	if err := os.MkdirAll(heapFilesPath, os.ModePerm); err != nil {
		return err
	}

	heapStemsPath := filepath.Join(heapPath, "stems")
	if err := os.MkdirAll(heapStemsPath, os.ModePerm); err != nil {
		return err
	}

	heapDerivationsPath := filepath.Join(heapPath, "derivations")
	if err := os.MkdirAll(heapDerivationsPath, os.ModePerm); err != nil {
		return err
	}

	heapContextsPath := filepath.Join(heapPath, "contexts")
	if err := os.MkdirAll(heapContextsPath, os.ModePerm); err != nil {
		return err
	}

	heapSecretsPath := filepath.Join(heapPath, "secrets")
	if err := os.MkdirAll(heapSecretsPath, os.ModePerm); err != nil {
		return err
	}

	rootsPath := filepath.Join(gardenPath, "roots")
	if err := os.MkdirAll(rootsPath, os.ModePerm); err != nil {
		return err
	}

	sproutsPath := filepath.Join(gardenPath, "sprouts")
	if err := os.MkdirAll(sproutsPath, os.ModePerm); err != nil {
		return err
	}

	return nil
}

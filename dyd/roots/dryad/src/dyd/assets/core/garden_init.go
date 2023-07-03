package core

import (
	"fmt"
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

	derivationsPath := filepath.Join(heapPath, "derivations")
	if err := os.MkdirAll(derivationsPath, os.ModePerm); err != nil {
		return err
	}

	contextsPath := filepath.Join(heapPath, "contexts")
	if err := os.MkdirAll(contextsPath, os.ModePerm); err != nil {
		return err
	}

	secretsPath := filepath.Join(heapPath, "secrets")
	if err := os.MkdirAll(secretsPath, os.ModePerm); err != nil {
		return err
	}

	shedPath := filepath.Join(gardenPath, "shed")
	if err := os.MkdirAll(shedPath, os.ModePerm); err != nil {
		return err
	}

	scopesPath := filepath.Join(shedPath, "scopes")
	if err := os.MkdirAll(scopesPath, os.ModePerm); err != nil {
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

	// write out type file
	typePath := filepath.Join(gardenPath, "type")

	typeFile, err := os.Create(typePath)
	if err != nil {
		return err
	}
	defer typeFile.Close()

	_, err = fmt.Fprint(typeFile, "garden")
	if err != nil {
		return err
	}

	err = typeFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

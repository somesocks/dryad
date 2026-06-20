package core

import (
	"dryad/internal/filepath"
	"dryad/internal/os"
)

func SproutInit(path string) error {
	dydPath := filepath.Join(path, "dyd")
	if err := os.Mkdir(dydPath, os.ModePerm); err != nil {
		return err
	}

	dependenciesPath := filepath.Join(dydPath, "dependencies")
	if err := os.Mkdir(dependenciesPath, os.ModePerm); err != nil {
		return err
	}

	traitsPath := filepath.Join(dydPath, "traits")
	if err := os.Mkdir(traitsPath, os.ModePerm); err != nil {
		return err
	}

	return nil
}

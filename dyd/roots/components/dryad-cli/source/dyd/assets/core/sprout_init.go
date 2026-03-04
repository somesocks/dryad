package core

import (
	"dryad/internal/os"
	stdos "os"
	"path/filepath"
)

func SproutInit(path string) error {
	dydPath := filepath.Join(path, "dyd")
	if err := os.MkdirAll(dydPath, stdos.ModePerm); err != nil {
		return err
	}

	dependenciesPath := filepath.Join(dydPath, "dependencies")
	if err := os.MkdirAll(dependenciesPath, stdos.ModePerm); err != nil {
		return err
	}

	traitsPath := filepath.Join(dydPath, "traits")
	if err := os.MkdirAll(traitsPath, stdos.ModePerm); err != nil {
		return err
	}

	return nil
}

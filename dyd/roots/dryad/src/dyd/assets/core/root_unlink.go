package core

import (
	"errors"
	"os"
	"path/filepath"
)

func RootUnlink(rootPath string, depPath string) error {
	if depPath == "" {
		return errors.New("missing path to dependency root")
	}

	depPath, err := filepath.Abs(depPath)
	if err != nil {
		return err
	}

	err = os.Remove(depPath)
	return err
}

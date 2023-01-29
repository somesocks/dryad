package core

import (
	"errors"
	"os"
	"path/filepath"
)

func RootAdd(rootPath string, depPath string, alias string) error {
	var err error
	if depPath == "" {
		return errors.New("missing path to dependency root")
	}

	depPath, err = filepath.Abs(depPath)
	if err != nil {
		return err
	}

	if alias == "" {
		alias = filepath.Base(depPath)
	}

	rootPath, err = RootPath(rootPath)
	if err != nil {
		return err
	}

	var rootsPath = filepath.Join(rootPath, "dyd", "roots")
	var aliasPath = filepath.Join(rootsPath, alias)

	var linkPath string
	linkPath, err = filepath.Rel(rootsPath, depPath)
	if err != nil {
		return err
	}

	err = os.Symlink(linkPath, aliasPath)
	if err != nil {
		return err
	}

	return nil
}

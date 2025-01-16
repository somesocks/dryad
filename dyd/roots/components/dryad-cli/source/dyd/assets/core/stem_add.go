package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"

	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func StemAdd(rootPath string, fingerprint string, alias string) error {
	var err error

	rootPath, err = RootPath(rootPath, "")
	if err != nil {
		return err
	}

	var depPath string
	depPath, err = HeapHasStem(rootPath, fingerprint)
	if err != nil {
		return err
	}
	if depPath == "" {
		return errors.New("heap missing stem with fingerprint " + fingerprint)
	}

	var stemsPath = filepath.Join(rootPath, "dyd", "dependencies")
	var aliasPath = filepath.Join(rootPath, "dyd", "dependencies", alias)

	var linkPath string
	linkPath, err = filepath.Rel(stemsPath, depPath)

	fmt.Println("StemAdd ", stemsPath, " ", linkPath, " ", depPath, " ", aliasPath)

	if err != nil {
		return err
	}

	err, _ = dydfs.RemoveAll(task.SERIAL_CONTEXT, aliasPath)
	if err != nil {
		return err
	}

	err = os.Symlink(linkPath, aliasPath)
	if err != nil {
		return err
	}

	return nil
}

package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"
)

func RootMove(sourcePath string, destPath string) error {

	// normalize the source path
	sourcePath, err := RootPath(sourcePath, "")
	if err != nil {
		return err
	}

	// copy the root to the new path
	err = RootCopy(sourcePath, destPath)
	if err != nil {
		return err
	}

	// replace references to the root
	err = RootReplace(sourcePath, destPath)
	if err != nil {
		return err
	}

	// delete the old root
	err, _ = dydfs.RemoveAll(task.SERIAL_CONTEXT, sourcePath)
	return err
}

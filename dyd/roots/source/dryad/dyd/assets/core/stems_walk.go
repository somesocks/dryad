package core

import (
	"os"
	"path/filepath"
)

func StemsWalk(path string, walkFn filepath.WalkFunc) error {
	var stemsPath, err = StemsPath(path)
	if err != nil {
		return err
	}

	var rootsDir *os.File
	rootsDir, err = os.Open(stemsPath)
	if err != nil {
		return err
	}

	var files []os.DirEntry
	files, err = rootsDir.ReadDir(0)
	if err != nil {
		return err
	}

	for _, file := range files {
		var fileInfo, fileInfoErr = file.Info()
		if fileInfoErr != nil {
			return fileInfoErr
		}
		err = walkFn(filepath.Join(stemsPath, fileInfo.Name()), fileInfo, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

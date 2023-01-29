package core

import (
	"os"
	"path/filepath"
)

func RootsWalk(path string, walkFn filepath.WalkFunc) error {
	var rootsPath, err = RootsPath(path)
	if err != nil {
		return err
	}

	var rootsDir *os.File
	rootsDir, err = os.Open(rootsPath)
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
		err = walkFn(filepath.Join(rootsPath, fileInfo.Name()), fileInfo, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

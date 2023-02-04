package core

import (
	"errors"
	"os"
	"path/filepath"
)

func GardenPath(path string) (string, error) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}

	var workingPath = path
	var flagPath = filepath.Join(workingPath, "dyd", "garden")
	var _, fileInfoErr = os.Stat(flagPath)

	for workingPath != "/" {

		if fileInfoErr == nil {
			return workingPath, nil
		}

		workingPath = filepath.Dir(workingPath)
		flagPath = filepath.Join(workingPath, "dyd", "garden")
		_, fileInfoErr = os.Stat(flagPath)
	}

	return "", errors.New("dyd garden path not found starting from " + path)
}

package core

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func RootPath(path string) (string, error) {
	var workingPath, err = filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	var mainPath = filepath.Join(workingPath, "dyd", "root")
	var _, fileInfoErr = os.Stat(mainPath)

	for workingPath != "/" {

		if fileInfoErr == nil {
			return workingPath, nil
		}

		workingPath = filepath.Dir(workingPath)
		mainPath = filepath.Join(workingPath, "dyd", "root")
		_, fileInfoErr = os.Stat(mainPath)
	}

	return "", errors.New("dyd root path not found for " + path)
}

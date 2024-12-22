package core

import (
	"errors"
	"os"
	"path/filepath"
)

func StemPath(path string) (string, error) {
	var working_path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}

	var dyd_path = filepath.Join(working_path, "dyd")
	var fileInfo, fileInfoErr = os.Stat(dyd_path)

	for working_path != "/" {

		if fileInfoErr == nil && fileInfo.IsDir() {
			return working_path, nil
		}

		working_path = filepath.Dir(working_path)
		dyd_path = filepath.Join(working_path, "dyd")
		fileInfo, fileInfoErr = os.Stat(dyd_path)
	}

	return "", errors.New("dyd stem path not found")
}

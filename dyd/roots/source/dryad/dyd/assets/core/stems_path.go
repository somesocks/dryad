package core

import (
	"errors"
	"os"
	"path/filepath"
)

func StemsPath(path string) (string, error) {
	var working_path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}

	var heap_path = filepath.Join(working_path, "dyd", "stems")
	var fileInfo, fileInfoErr = os.Stat(heap_path)

	for working_path != "/" {

		if fileInfoErr == nil && fileInfo.IsDir() {
			return heap_path, nil
		}

		working_path = filepath.Dir(working_path)
		heap_path = filepath.Join(working_path, "dyd", "stems")
		fileInfo, fileInfoErr = os.Stat(heap_path)
	}

	return "", errors.New("dyd stems path not found")
}

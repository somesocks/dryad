package core

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func GardenPath(path string) (string, error) {
	var working_path, err = filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	var heap_path = filepath.Join(working_path, "dyd", "heap")
	var fileInfo, fileInfoErr = os.Stat(heap_path)

	for working_path != "/" {

		if fileInfoErr == nil && fileInfo.IsDir() {
			return working_path, nil
		}

		working_path = filepath.Dir(working_path)
		heap_path = filepath.Join(working_path, "dyd", "heap")
		fileInfo, fileInfoErr = os.Stat(heap_path)
	}

	return "", errors.New("dyd garden path not found")
}

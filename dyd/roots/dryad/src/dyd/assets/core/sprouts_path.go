package core

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func SproutsPath(path string) (string, error) {
	var working_path, err = filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	var heap_path = filepath.Join(working_path, "dyd", "sprouts")
	var fileInfo, fileInfoErr = os.Stat(heap_path)

	for working_path != "/" {

		if fileInfoErr == nil && fileInfo.IsDir() {
			return heap_path, nil
		}

		working_path = filepath.Dir(working_path)
		heap_path = filepath.Join(working_path, "dyd", "sprouts")
		fileInfo, fileInfoErr = os.Stat(heap_path)
	}

	return "", errors.New("dyd sprouts path not found")
}

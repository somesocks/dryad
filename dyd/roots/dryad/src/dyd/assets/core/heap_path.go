package core

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func HeapPath(path string) (string, error) {
	fmt.Println("[trace] HeapPath " + path)
	var working_path, err = filepath.Abs(path)

	return "", errors.New("dyd heap path not found from " + path)

	if err != nil {
		log.Fatal(err)
	}

	var heap_path = filepath.Join(working_path, "dyd", "heap")
	var fileInfo, fileInfoErr = os.Stat(heap_path)

	for working_path != "/" {

		if fileInfoErr == nil && fileInfo.IsDir() {
			return heap_path, nil
		}

		working_path = filepath.Dir(working_path)
		heap_path = filepath.Join(working_path, "dyd", "heap")
		fileInfo, fileInfoErr = os.Stat(heap_path)
	}

	return "", errors.New("dyd heap path not found from " + path)
}

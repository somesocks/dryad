package core

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func RootPath(path string) (string, error) {
	var working_path, err = filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	var traits_path = filepath.Join(working_path, "dyd", "traits")
	var fileInfo, fileInfoErr = os.Stat(traits_path)

	for working_path != "/" {

		if fileInfoErr == nil && fileInfo.IsDir() {
			return working_path, nil
		}

		working_path = filepath.Dir(working_path)
		traits_path = filepath.Join(working_path, "dyd", "traits")
		fileInfo, fileInfoErr = os.Stat(traits_path)
	}

	return "", errors.New("dyd root path not found")
}

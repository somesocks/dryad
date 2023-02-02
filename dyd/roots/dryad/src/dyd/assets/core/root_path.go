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

	var main_path = filepath.Join(working_path, "dyd", "main")
	var _, fileInfoErr = os.Stat(main_path)

	for working_path != "/" {

		if fileInfoErr == nil {
			return working_path, nil
		}

		working_path = filepath.Dir(working_path)
		main_path = filepath.Join(working_path, "dyd", "main")
		_, fileInfoErr = os.Stat(main_path)
	}

	return "", errors.New("dyd root path not found for " + path)
}

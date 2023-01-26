package core

import (
	"errors"
	"io/fs"
	"os"
)

func fileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

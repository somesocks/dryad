package fs2

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func _mkDir(path string, permissions fs.FileMode) error {
	parentPath := filepath.Dir(path)
	parentInfo, err := os.Lstat(parentPath)
	if errors.Is(err, fs.ErrNotExist) {
		err := _mkDir(parentPath, permissions)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !parentInfo.IsDir() {
		return errors.New("parent path not directory")
	}

	parentMode := parentInfo.Mode()
	if parentMode&0o200 != 0o200 {
		err = os.Chmod(parentPath, parentMode|200)
		if err != nil {
			return err
		}

		err = os.Mkdir(path, permissions)
		if err != nil {
			return err
		}

		err = os.Chmod(parentPath, parentMode)
		if err != nil {
			return err
		}

	} else {

		err = os.Mkdir(path, permissions)
		if err != nil {
			return err
		}

	}

	return nil
}

func MkDir(path string, permissions fs.FileMode) error {
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if info != nil {
		if info.IsDir() {
			return nil
		} else {
			return errors.New("path exists as file")
		}
	} else {
		return _mkDir(path, permissions)
	}

}

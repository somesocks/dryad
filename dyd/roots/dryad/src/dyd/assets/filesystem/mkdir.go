package fs2

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func MkDir(path string, permissions fs.FileMode) error {
	// check if file exists
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if info != nil {
		if info.IsDir() {
			err = os.Chmod(path, permissions)
			if err != nil {
				return err
			}

			return nil
		} else {
			return errors.New("path exists as file")
		}
	} else {
		parentPath := filepath.Dir(path)

		err = MkDir(parentPath, permissions)
		if err != nil {
			return err
		}

		parentInfo, err := os.Lstat(parentPath)
		if err != nil {
			return err
		}

		parentMode := parentInfo.Mode()

		if parentMode&0o200 != 0o200 {
			err = os.Chmod(parentPath, parentMode|0o200)
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
}

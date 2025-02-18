package fs2

import (
	"dryad/task"

	"errors"
	"io/fs"
	"os"
	// "path/filepath"
)

type MkdirRequest struct {
	Path string
	Permissions fs.FileMode
}

type MkdirResult = MkdirRequest

func Mkdir2(ctx *task.ExecutionContext, req MkdirRequest) (error, *MkdirResult) {
	var res = MkdirResult{
		Path: req.Path,
		Permissions: req.Permissions,
	}
	var err error


	err = os.Mkdir(req.Path, req.Permissions)

	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return err, nil
		}
		// if the file already exists, check to see if it's a directory

		var info fs.FileInfo
		info, err = os.Lstat(req.Path)
		if err != nil {
			return err, nil
		}
	
		if info.IsDir() {
			err = os.Chmod(req.Path, req.Permissions)
			if err != nil {
				return err, nil
			}

			return nil, &res
		} else {
			return errors.New("path exists as file"), nil
		}

	}

	return nil, &res
}

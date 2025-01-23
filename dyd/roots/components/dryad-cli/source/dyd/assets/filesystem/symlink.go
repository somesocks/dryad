package fs2

import (
	// "errors"
	"io/fs"
	"os"
	"path/filepath"

	"dryad/task"
)

type SymlinkRequest struct {
	Path string
	Target string
}

func Symlink(ctx *task.ExecutionContext, request SymlinkRequest) (error, SymlinkRequest) {
	var parentPath string = filepath.Dir(request.Path)
	var parentFile *os.File
	var parentLock FileLock
	var err error

	// create a file descriptor to the parent directory,
	// so that we can grab a lock on it
	parentFile, err = os.OpenFile(parentPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err, request
	} 

	parentLock = newFileLock(parentFile)

	// grab a lock on the parent
	err = parentLock.Lock()
	if err != nil {
		return err, request
	}
	defer parentLock.Unlock()

	// grab the fileinfo for the parent
	var parentInfo fs.FileInfo
	parentInfo, err = os.Lstat(parentPath)
	if err != nil {
		return err, request
	}

	// if the parent permissions are not writable by the current user,
	// temporarily set the permissions to writable
	var parentMode fs.FileMode = parentInfo.Mode()
	if (parentMode & 0400) != (parentMode & 0777) {
		err = os.Chmod(parentPath, parentMode & 0400)
		if err != nil {
			return err, request
		}
		defer os.Chmod(parentPath, parentMode)
	}

	// create the symlink and return
	err = os.Symlink(request.Target, request.Path)

	return err, request
}

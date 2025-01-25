package fs2

import (
	// "errors"
	"io/fs"
	"os"
	"errors"
	"path/filepath"

	"dryad/task"
)

type SymlinkRequest struct {
	Path string
	Target string
}

func fileExists(filename string) (error, bool) {
	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false 
		} else {
			return err, false
		}
	} else {
		return nil, true
	}
}

func Symlink(ctx *task.ExecutionContext, request SymlinkRequest) (error, SymlinkRequest) {
	var parentPath string = filepath.Dir(request.Path)
	var parentFile *os.File
	var parentLock FileLock
	var err error

	// create a file descriptor to the parent directory,
	// so that we can grab a lock on it
	parentFile, err = os.Open(parentPath)
	if err != nil {
		return err, request
	}
	defer parentFile.Close()


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
	var parentPerms = parentInfo.Mode().Perm()
	if (parentPerms | 0o400) != parentPerms {
		err = os.Chmod(parentPath, parentPerms | 0o400)
		if err != nil {
			return err, request
		}
		defer os.Chmod(parentPath, parentPerms)
	}

	// check if the symlink already exists
	err, exists := fileExists(request.Path)
	if err != nil {
		return err, request
	}

	// remove the symlink if it already exists
	if exists {
		err = os.Remove(request.Path)
		if err != nil {
			return err, request
		}	
	}	

	// create the symlink and return
	err = os.Symlink(request.Target, request.Path)

	return err, request
}

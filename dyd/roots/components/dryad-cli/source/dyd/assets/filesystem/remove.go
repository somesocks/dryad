package fs2

import (
	"io/fs"
	"os"
	"path/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

var remove = func (ctx *task.ExecutionContext, path string) (error, any) {
	var parentPath string = filepath.Dir(path)
	var err error

	// grab the fileinfo for the parent
	var parentInfo fs.FileInfo
	parentInfo, err = os.Lstat(parentPath)
	if err != nil {
		zlog.Error().
			Str("path", path).
			Err(err).
			Msg("dydfs.remove - get parent info")
		return err, nil
	}

	// if the parent permissions are not writable by the current user,
	// temporarily set the permissions to writable
	var parentPerms = parentInfo.Mode()
	if (parentPerms | 0o200) != parentPerms {
		err, _ = Chmod(
			ctx,
			ChmodRequest{
				Path: parentPath,
				Mode: parentPerms | 0o200,
				SkipLock: true,
			},
		)
		if err != nil {
			zlog.Error().
				Str("path", path).
				Err(err).
				Msg("dydfs.remove - parent chmod")
			return err, nil
		}
		defer Chmod(
			ctx,
			ChmodRequest{
				Path: parentPath,
				Mode: parentPerms,
				SkipLock: true,
			},
		)
	}

	err = os.Remove(path)
	if err != nil {
		zlog.Error().
			Str("path", path).
			Err(err).
			Msg("dydfs.remove - os.remove")
		return err, nil
	}

	return nil, nil
}

var remove2 = WithFileLock(
	remove,
	func (ctx *task.ExecutionContext, path string) (error, string) {
		return nil, filepath.Dir(path)
	},
)

var Remove = remove2

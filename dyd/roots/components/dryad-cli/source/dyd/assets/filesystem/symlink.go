package fs2

import (
	// "errors"
	"io/fs"
	"os"
	// "errors"
	"path/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"

	"math/rand"
	"fmt"
)

type SymlinkRequest struct {
	Path string
	Target string
}


var Symlink task.Task[SymlinkRequest, SymlinkRequest] = func () task.Task[SymlinkRequest, SymlinkRequest] {

	// var fileExists = func (filename string) (error, bool) {
	// 	_, err := os.Stat(filename)
	// 	if err != nil {
	// 		if errors.Is(err, fs.ErrNotExist) {
	// 			return nil, false 
	// 		} else {
	// 			return err, false
	// 		}
	// 	} else {
	// 		return nil, true
	// 	}
	// }

	var symlink = func (ctx *task.ExecutionContext, request SymlinkRequest) (error, SymlinkRequest) {
		var parentPath string = filepath.Dir(request.Path)
		var err error
	
		// grab the fileinfo for the parent
		var parentInfo fs.FileInfo
		parentInfo, err = os.Lstat(parentPath)
		if err != nil {
			zlog.Error().
				Str("path", request.Path).
				Err(err).
				Msg("dydfs.symlink - get parent info")
			return err, request
		}
	
		// if the parent permissions are not writable by the current user,
		// temporarily set the permissions to writable
		var parentPerms = parentInfo.Mode()
		if (parentPerms | 0o770) != parentPerms {
			err, _ = Chmod(
				ctx,
				ChmodRequest{
					Path: parentPath,
					Mode: parentPerms | 0o770,
					SkipLock: true,
				},
			)
			if err != nil {
				zlog.Error().
					Str("path", request.Path).
					Err(err).
					Msg("dydfs.symlink - parent chmod")
				return err, request
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
	
		var tempPath string = fmt.Sprintf(
			"%s.tmp.%x",
			request.Path,
			rand.Int63(),
		)
	
		err = os.Symlink(request.Target, tempPath)
		if err != nil {
			zlog.Error().
				Str("path", request.Path).
				Str("tempPath", tempPath).
				Err(err).
				Msg("dydfs.symlink - create temporary symlink")
			return err, request
		}

		err = os.Rename(tempPath, request.Path)
		if err != nil {
			zlog.Error().
				Str("path", request.Path).
				Str("tempPath", tempPath).
				Err(err).
				Msg("dydfs.symlink - move temporary symlink")
			defer os.Remove(tempPath)
			return err, request
		}

		return nil, request
	}
	
	symlink = WithFileLock(
		symlink,
		func (ctx *task.ExecutionContext, request SymlinkRequest) (error, string) {
			return nil, filepath.Dir(request.Path)
		},
	)

	symlink = task.Series2(
		func (ctx *task.ExecutionContext, req SymlinkRequest) (error, SymlinkRequest) {
			zlog.Trace().
				Str("path", req.Path).
				Msg("dydfs.symlink")
			return nil, req
		},
		symlink,
	)

	return symlink
}()
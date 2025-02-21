package fs2

import (
	"dryad/task"

	"errors"
	"io/fs"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type MkdirRequest struct {
	Path string
	Permissions fs.FileMode
}

type MkdirResult = MkdirRequest

var Mkdir2 = func () task.Task[MkdirRequest, *MkdirResult] {

	var mkdir2 = func(ctx *task.ExecutionContext, req MkdirRequest) (error, *MkdirResult) {
		var res = MkdirResult{
			Path: req.Path,
			Permissions: req.Permissions,
		}
		var err error
	
		var parentPath string = filepath.Dir(req.Path)
	
		// grab the fileinfo for the parent
		var parentInfo fs.FileInfo
		parentInfo, err = os.Lstat(parentPath)
		if err != nil {
			zlog.Error().
				Str("path", req.Path).
				Err(err).
				Msg("dydfs.mkdir2 - get parent info")
			return err, &res
		}
	
		// if the parent permissions are not writable by the current user,
		// temporarily set the permissions to writable
		var parentPerms = parentInfo.Mode()
		zlog.Trace().
			Str("parentPath", req.Path).
			Str("parentPerms", parentPerms.String()).
			Str("newParentPerms", (parentPerms | 0o770).String()).
			Msg("dydfs.mkdir2 - parent perms")

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
					Str("path", req.Path).
					Err(err).
					Msg("dydfs.mkdir2 - parent chmod")
				return err, &res
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
	
		err = os.Mkdir(req.Path, req.Permissions)
		if err == nil {
			return nil, &res
		} else if errors.Is(err, fs.ErrExist) {
			// if the file already exists, check to see if it's a directory
			var info fs.FileInfo
			info, err = os.Lstat(req.Path)
			if err != nil {
				zlog.Error().
					Err(err).
					Msg("dydfs.Mkdir2 lstat error")
				return err, nil
			}
		
			if info.IsDir() {
				newMode := (info.Mode() & 0xFFFFFE00) | req.Permissions
				zlog.Trace().
					Str("orig_perms", info.Mode().String()).
					Str("new_perms", newMode.String()).
					Msg("dydfs.Mkdir2 chmod")
	
				err, _ = Chmod(ctx, ChmodRequest{ Path: req.Path, Mode: newMode })
				if err != nil {
					zlog.Error().
					Err(err).
					Msg("dydfs.Mkdir2 chmod error")
					return err, nil
				}
	
				return nil, &res
			} else {
				return errors.New("path exists as file"), nil
			}		
		} else {
			return err, nil
		}
	
	}
	
	mkdir2 = WithFileLock(
		mkdir2,
		func (ctx *task.ExecutionContext, req MkdirRequest) (error, string) {
			return nil, filepath.Dir(req.Path)
		},
	)
	
	mkdir2 = task.Series2(
		func (ctx *task.ExecutionContext, req MkdirRequest) (error, MkdirRequest) {
			zlog.Trace().
				Str("path", req.Path).
				Str("mode", req.Permissions.String()).
				Msg("dydfs.mkdir2")
			return nil, req
		},
		mkdir2,
	)

	return mkdir2
}()
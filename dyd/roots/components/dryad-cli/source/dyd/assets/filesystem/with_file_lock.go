
package fs2

import (
	"dryad/task"

	"os"
	"io/fs"
	"path/filepath"
	// "fmt"

	"runtime/debug"
	zlog "github.com/rs/zerolog/log"
)

func WithFileLock[A any, B any] (
	baseTask task.Task[A, B],
	pathFunc task.Task[A, string],
) task.Task[A, B] {
	var wrapper task.Task[A, B] = func (ctx *task.ExecutionContext, req A) (error, B) {
		var path string
		var err error
		var res B

		err, path = pathFunc(ctx, req)
		if err != nil {
			zlog.Error().
				Str("path", path).
				Err(err).
				Msg("dydfs.WithFileLock - get path")
			return err, res
		}
		zlog.Trace().
			Str("path", path).
			Msg("dydfs.WithFileLock - path")

		// if no path is returned, skip locking
		if path != "" {
			// create a file descriptor,
			// so that we can grab a lock on it
			file, err := os.OpenFile(path, os.O_RDONLY, 0o777)
			if err != nil {
				zlog.Error().
					Str("path", path).
					Err(err).
					Msg("dydfs.WithFileLock - get file descriptor")
				debug.PrintStack()

				var parentPath = filepath.Dir(path)

				// grab the fileinfo for the parent
				var parentInfo fs.FileInfo
				parentInfo, err = os.Lstat(parentPath)
				if err != nil {
					zlog.Error().
						Str("path", parentPath).
						Err(err).
						Msg("dydfs.WithFileLock - get parent info")
					return err, res
				}

				zlog.Trace().
					Str("path", parentPath).
					Str("perms", parentInfo.Mode().String()).
					Msg("dydfs.WithFileLock - parent perms")

				return err, res
			}
			defer file.Close()

			// create the lock
			lock := newFileLock(file)

			// grab a lock on the parent
			err = lock.Lock()
			zlog.Trace().
				Str("path", path).
				Msg("dydfs.WithFileLock - grab lock")
			if err != nil {
				zlog.Error().
					Str("path", path).
					Err(err).
					Msg("dydfs.WithFileLock - grab lock error")
				return err, res
			}
			defer func () {
				zlog.Trace().
					Str("path", path).
					Msg("dydfs.WithFileLock - release lock")
				lock.Unlock()
			}()
		}

		err, res = baseTask(ctx, req)
		return err, res
	}

	return wrapper
}
package fs2

import (
	"dryad/task"

	"errors"
	"io/fs"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var PartialEvalSymlinks = func () task.Task[string, string] {

	var partialEvalSymlinks = func(ctx *task.ExecutionContext, path string) (error, string) {		
		var existsPath string
		var nonExistsPath string
		var err error

		existsPath, err = filepath.Abs(path)
		if err != nil {
			return err, ""
		}

		for {
			if existsPath == "." {
				break
			}

			_, err = os.Lstat(existsPath)
			if err == nil {
				break
			} else if errors.Is(err, fs.ErrNotExist) {
				nonExistsPath = filepath.Join(filepath.Base(existsPath), nonExistsPath)
				existsPath = filepath.Dir(existsPath)
			} else {
				return err, ""
			}
	
		}

		if existsPath == "." {
			return nil, path
		}

		existsPath, err = filepath.EvalSymlinks(existsPath)
		if err != nil {
			return err, ""
		}

		return nil, filepath.Join(existsPath, nonExistsPath)
	}

	partialEvalSymlinks = task.Series2(
		func (ctx *task.ExecutionContext, path string) (error, string) {
			zlog.Trace().
				Str("path", path).
				Msg("dydfs.PartialEvalSymlinks")
			return nil, path
		},
		partialEvalSymlinks,
	)

	return partialEvalSymlinks
}()
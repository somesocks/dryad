package fs2

import (
	"dryad/task"

	// "errors"
	"io/fs"
	"os"
	// "path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type ChmodRequest struct {
	Path string
	Mode fs.FileMode
	SkipLock bool
}

type ChmodResult = ChmodRequest

var Chmod = func () task.Task[ChmodRequest, ChmodResult] {

	var chmod = func(ctx *task.ExecutionContext, req ChmodRequest) (error, ChmodResult) {		
		var res = ChmodResult{
			Path: req.Path,
			Mode: req.Mode,
		}
		var err error
		err = os.Chmod(req.Path, req.Mode)
		return err, res
	}

	chmod = WithFileLock(
		chmod,
		func (ctx *task.ExecutionContext, req ChmodRequest) (error, string) {
			if req.SkipLock { return nil, "" }
			return nil, req.Path
		},
	)
		
	chmod = task.Series2(
		func (ctx *task.ExecutionContext, req ChmodRequest) (error, ChmodRequest) {
			zlog.Trace().
				Str("path", req.Path).
				Str("mode", req.Mode.String()).
				Msg("dydfs.Chmod")
			return nil, req
		},
		chmod,
	)

	return chmod
}()
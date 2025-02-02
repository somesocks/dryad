package fs2

import (
	// "io/fs"
	"os"
	"path/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

func RemoveAll(ctx *task.ExecutionContext, path string) (error, any) {
	zlog.Trace().
		Str("path", path).
		Msg("dryad/filesystem/RemoveAll")

	_, err := os.Lstat(path);
	if err != nil {
		zlog.Trace().
			Err(err).
			Bool("existsErr", os.IsNotExist(err)).
			Msg("dryad/filesystem/RemoveAll path err")
		// if the path does not exist, silently return
		if os.IsNotExist(err) {
			return nil, nil
		} else {
			return err, nil
		}	
	}

	// walk through the filesystem and fix any permissions problems,
	// if you can
	err = DFSWalk3(
		ctx,
		Walk5Request{
			BasePath:    path,
			Path:        path,
			VPath:       path,
			ShouldCrawl: func(ctx *task.ExecutionContext, node Walk5Node) (error, bool) {
				// don't crawl symlinks
				var shouldCrawl bool = !(node.Info.Mode()&os.ModeSymlink == os.ModeSymlink)

				zlog.Trace().
					Str("path", node.Path).
					Str("vpath", node.VPath).
					Bool("shouldCrawl", shouldCrawl).
					Bool("isSymLink", node.Info.Mode()&os.ModeSymlink == os.ModeSymlink).
					Msg("dryad/filesystem/RemoveAll/ShouldCrawl")

				return nil, shouldCrawl
			},
			ShouldMatch: func(ctx *task.ExecutionContext, node Walk5Node) (error, bool) {
				var shouldMatch bool = true
				zlog.Trace().
					Str("path", node.Path).
					Str("vpath", node.VPath).
					Bool("shouldMatch", shouldMatch).
					Msg("dryad/filesystem/RemoveAll/ShouldMatch")
				return nil, shouldMatch
			},
			OnMatch: func(ctx *task.ExecutionContext, node Walk5Node) (error, any) {
				zlog.Trace().
					Str("path", node.Path).
					Str("vpath", node.VPath).
					Msg("dryad/filesystem/RemoveAll/OnMatch")

				parentInfo, err := os.Lstat(filepath.Dir(node.Path))
				if err != nil {
					return err, nil
				}

				if parentInfo.Mode()&0o200 != 0o200 {
					err := os.Chmod(filepath.Dir(node.Path), parentInfo.Mode()|0o200)
					if err != nil {
						return err, nil
					}
				}

				err = os.Remove(node.Path)
				if err != nil {
					return err, nil
				}

				return nil, nil
			},
		},
	)
	if err != nil {
		zlog.Trace().
			Err(err).
			Bool("existsErr", os.IsNotExist(err)).
			Msg("dryad/filesystem/RemoveAll/DFSWalk err")
		return err, nil
	}

	return nil, nil
}

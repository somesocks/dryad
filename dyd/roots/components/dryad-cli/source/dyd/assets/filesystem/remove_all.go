package fs2

import (
	// "io/fs"
	"os"
	// "path/filepath"

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

	shouldWalk := func(ctx *task.ExecutionContext, node Walk6Node) (error, bool) {
		isSymlink := node.Info.Mode()&os.ModeSymlink == os.ModeSymlink
		shouldWalk := !isSymlink
 
		zlog.Trace().
			Str("path", node.Path).
			Str("vpath", node.VPath).
			Bool("isSymlink", isSymlink).
			Msg("dryad/filesystem/RemoveAll/shouldWalk")

		return nil, shouldWalk
	}

	onPreMatch := func(ctx *task.ExecutionContext, node Walk6Node) (error, any) {
		isDir := node.Info.IsDir()
		isWritable := node.Info.Mode()&0o200 == 0o200

		zlog.Trace().
			Str("path", node.Path).
			Str("vpath", node.VPath).
			Bool("isDir", isDir).
			Bool("isWritable", isWritable).
			Msg("dryad/filesystem/RemoveAll/onPreMatch")

		if isDir && !isWritable {
			err := os.Chmod(node.Path, node.Info.Mode()|0o200)

			zlog.Trace().
				Str("path", node.Path).
				Str("vpath", node.VPath).
				Err(err).
				Msg("dryad/filesystem/RemoveAll/onPreMatch chmod")

			if err != nil {
				return err, nil
			}
		}

		return nil, nil
	}

	onPostMatch := func(ctx *task.ExecutionContext, node Walk6Node) (error, any) {
		isWritable := node.Info.Mode()&0o200 != 0o200
		isDir := node.Info.IsDir()

		err = os.Remove(node.Path)

		zlog.Trace().
			Str("path", node.Path).
			Str("vpath", node.VPath).
			Bool("isWritable", isWritable).
			Bool("isDir", isDir).
			Err(err).
			Msg("dryad/filesystem/RemoveAll/onPostMatch remove")

		return err, nil
	}

	err, _ = Walk6(
		ctx,
		Walk6Request{
			BasePath:    path,
			Path:        path,
			VPath:       path,
			ShouldWalk: shouldWalk,
			OnPreMatch: onPreMatch,
			OnPostMatch: onPostMatch,
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

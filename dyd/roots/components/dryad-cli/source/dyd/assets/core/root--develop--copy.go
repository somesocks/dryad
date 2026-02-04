package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"

	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type rootDevelopCopyOptions struct {
	ApplyIgnore bool
}

func rootDevelop_copyDir(
	ctx *task.ExecutionContext,
	srcPath string,
	destPath string,
	options rootDevelopCopyOptions,
) error {
	var err error

	if ctx == nil {
		ctx = task.DEFAULT_CONTEXT
	}
	ctx = &task.ExecutionContext{
		ConcurrencyChannel: ctx.ConcurrencyChannel,
	}

	srcPath, err = filepath.Abs(srcPath)
	if err != nil {
		return err
	}

	info, err := os.Lstat(srcPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("rootDevelop_copyDir: source is not a directory: %s", srcPath)
	}

	err = os.MkdirAll(destPath, info.Mode())
	if err != nil {
		return err
	}

	shouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		if node.Info == nil {
			return nil, false
		}
		if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil, false
		}
		if !node.Info.IsDir() {
			return nil, false
		}

		if options.ApplyIgnore {
			parentDir := filepath.Dir(node.VPath)
			err, matcher := readDydIgnore(ctx, DydIgnoreRequest{
				BasePath: srcPath,
				Path:     parentDir,
			})
			if err != nil {
				return err, false
			}

			err, match := matcher.Match(dydfs.NewGlobPath(node.VPath, true))
			if err != nil {
				return err, false
			}
			if match {
				return nil, false
			}
		}

		return nil, true
	}

	onCopy := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		relPath, err := filepath.Rel(srcPath, node.VPath)
		if err != nil {
			return err, nil
		}
		if relPath == "." {
			return nil, nil
		}

		if options.ApplyIgnore {
			parentDir := filepath.Dir(node.VPath)
			err, matcher := readDydIgnore(ctx, DydIgnoreRequest{
				BasePath: srcPath,
				Path:     parentDir,
			})
			if err != nil {
				return err, nil
			}

			err, match := matcher.Match(dydfs.NewGlobPath(node.VPath, node.Info.IsDir()))
			if err != nil {
				return err, nil
			}
			if match {
				zlog.Trace().
					Str("path", node.VPath).
					Msg("rootDevelop_copyDir ignored")
				return nil, nil
			}
		}

		dest := filepath.Join(destPath, relPath)

		mode := node.Info.Mode()
		switch {
		case mode.IsDir():
			return os.MkdirAll(dest, mode), nil
		case mode&os.ModeSymlink == os.ModeSymlink:
			linkTarget, err := os.Readlink(node.Path)
			if err != nil {
				return err, nil
			}
			return os.Symlink(linkTarget, dest), nil
		case mode.IsRegular():
			return rootDevelop_copyFile(node.Path, dest, mode), nil
		default:
			return fmt.Errorf("rootDevelop_copyDir: unsupported file type: %s", node.Path), nil
		}
	}

	err, _ = dydfs.Walk6(
		ctx,
		dydfs.Walk6Request{
			BasePath:   srcPath,
			Path:       srcPath,
			VPath:      srcPath,
			ShouldWalk: shouldWalk,
			OnPreMatch: onCopy,
		},
	)
	return err
}

func rootDevelop_copyFile(srcPath string, destPath string, mode fs.FileMode) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return destFile.Chmod(mode)
}

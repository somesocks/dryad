package fs2

import (
	"dryad/internal/os"
	"dryad/task"
)

func RemoveAll(ctx *task.ExecutionContext, path string) (error, any) {
	_, err := os.Lstat(path)
	if err != nil {
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
		return nil, shouldWalk
	}

	onPreMatch := func(ctx *task.ExecutionContext, node Walk6Node) (error, any) {
		isDir := node.Info.IsDir()
		isWritable := node.Info.Mode()&0o200 == 0o200

		if isDir && !isWritable {
			err := os.Chmod(node.Path, node.Info.Mode()|0o200)
			if err != nil {
				return err, nil
			}
		}

		return nil, nil
	}

	onPostMatch := func(ctx *task.ExecutionContext, node Walk6Node) (error, any) {
		err = os.Remove(node.Path)
		return err, nil
	}

	err, _ = Walk6(
		ctx,
		Walk6Request{
			BasePath:    path,
			Path:        path,
			VPath:       path,
			ShouldWalk:  shouldWalk,
			OnPreMatch:  onPreMatch,
			OnPostMatch: onPostMatch,
		},
	)

	if err != nil {
		return err, nil
	}

	return nil, nil
}

package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	"io/fs"
	"path/filepath"
)

func RootRequirementsWalk(path string, walkFn func(path string, info fs.FileInfo) error) error {
	path, err := RootPath(path, "")
	if err != nil {
		return err
	}

	requirementsPath := filepath.Join(path, "dyd", "requirements")

	requirementsExists, err := fileExists(requirementsPath)
	if err != nil {
		return err
	}

	// if requirements doesn't exist, do nothing
	if !requirementsExists {
		return nil
	}

	err, _ = fs2.BFSWalk3(
		task.SERIAL_CONTEXT,
		fs2.Walk5Request{
			Path:     requirementsPath,
			VPath:    requirementsPath,
			BasePath: requirementsPath,
			ShouldCrawl: func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
				return nil, node.Path == node.BasePath
			},
			ShouldMatch: func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
				return nil, node.Path != node.BasePath
			},
			OnMatch: func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
				return walkFn(node.Path, node.Info), nil
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

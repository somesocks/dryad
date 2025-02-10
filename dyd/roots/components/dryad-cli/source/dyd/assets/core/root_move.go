package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"
)

type RootMoveRequest struct {
	Garden *SafeGardenReference
	SourcePath string
	DestPath string
}

func RootMove(ctx *task.ExecutionContext, req RootMoveRequest) (error, any) {
	var sourcePath string = req.SourcePath
	var destPath string = req.DestPath

	// normalize the source path
	sourcePath, err := RootPath(sourcePath, "")
	if err != nil {
		return err, nil
	}

	// copy the root to the new path
	err, _ = RootCopy(
		ctx,
		RootCopyRequest{
			Garden: req.Garden,
			SourcePath: sourcePath,
			DestPath: destPath,
		},
	)
	if err != nil {
		return err, nil
	}

	// replace references to the root
	err = RootReplace(
		RootReplaceRequest{
			Garden: req.Garden,
			SourcePath: sourcePath,
			DestPath: destPath,
		},
	)
	if err != nil {
		return err, nil
	}

	// delete the old root
	err, _ = dydfs.RemoveAll(task.SERIAL_CONTEXT, sourcePath)
	return err, nil
}

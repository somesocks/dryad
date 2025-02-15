package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"
	"fmt"
)

type rootMoveRequest struct {
	Source *SafeRootReference
	Dest *UnsafeRootReference
}

func rootMove(ctx *task.ExecutionContext, req rootMoveRequest) (error, *SafeRootReference) {
	var sourcePath string = req.Source.BasePath
	var err error

	// check that source and destination are within the same garden
	if req.Source.Roots.Garden.BasePath != req.Dest.Roots.Garden.BasePath {
		return fmt.Errorf("source and destination roots are not in same garden"), nil
	}

	var newRoot *SafeRootReference

	// copy the root to the new path
	err, newRoot = req.Source.Copy(
		ctx,
		RootCopyRequest{
			Dest: req.Dest,
		},
	)
	if err != nil {
		return err, nil
	}

	// replace references to the root
	err = RootReplace(
		RootReplaceRequest{
			Source: req.Source,
			Dest: newRoot,
		},
	)
	if err != nil {
		return err, nil
	}

	// delete the old root
	err, _ = dydfs.RemoveAll(task.SERIAL_CONTEXT, sourcePath)
	return err, nil
}

type RootMoveRequest struct {
	Dest *UnsafeRootReference
}

func (root *SafeRootReference) Move(ctx *task.ExecutionContext, req RootMoveRequest) (error) {
	err, _ := rootMove(
		ctx,
		rootMoveRequest{
			Source: root,
			Dest: req.Dest,
		},
	)
	return err
}
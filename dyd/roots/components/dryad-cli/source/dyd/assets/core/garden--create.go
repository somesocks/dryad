package core

import (
	"fmt"
	"os"
	"path/filepath"

	task "dryad/task"
)

type gardenCreateRequest struct {
	BasePath string
}

func gardenPrepareRequest(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path, err := filepath.Abs(req.BasePath)
	if err != nil {
		return err, req
	}
	req.BasePath = path
	return nil, req
}

func gardenCreateBase(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeap(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapFiles(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "files")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapStems(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "stems")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapDerivations(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "derivations")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapContexts(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "contexts")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapSecrets(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "secrets")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateShed(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "shed")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateShedScopes(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "shed", "scopes")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateRoots(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "roots")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateSprouts(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "sprouts")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateTypeFile(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	// write out type file
	typePath := filepath.Join(req.BasePath, "dyd", "type")

	typeFile, err := os.Create(typePath)
	if err != nil {
		return err, req
	}
	defer typeFile.Close()

	_, err = fmt.Fprint(typeFile, "garden")
	if err != nil {
		return err, req
	}

	return nil, req
}

var gardenCreate = task.Series4(
	gardenPrepareRequest,
	gardenCreateBase,
	task.Parallel5(
		task.Series3(
			gardenCreateHeap,
			task.Parallel5(
				gardenCreateHeapFiles,
				gardenCreateHeapStems,
				gardenCreateHeapDerivations,
				gardenCreateHeapContexts,
				gardenCreateHeapSecrets,
			),
			func (
				ctx *task.ExecutionContext,
				res task.Tuple5[gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest],
			) (error, gardenCreateRequest) {
				return nil, res.A
			},
		),
		task.Series2(
			gardenCreateShed,
			gardenCreateShedScopes,
		),
		gardenCreateRoots,
		gardenCreateSprouts,
		gardenCreateTypeFile,
	),
	func (
		ctx *task.ExecutionContext,
		res task.Tuple5[gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest],
	) (error, gardenCreateRequest) {
		return nil, res.A
	},
)

func (ug *UnsafeGardenReference) Create(ctx *task.ExecutionContext) (error, *SafeGardenReference) {
	err, _ := gardenCreate(
		ctx,
		gardenCreateRequest{
			BasePath: ug.BasePath,
		},
	)

	if err != nil {
		return err, nil
	}

	err, safeGardenRef := ug.Resolve(ctx)
	return err, safeGardenRef
}
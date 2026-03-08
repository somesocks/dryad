package core

import (
	"dryad/internal/os"
	"fmt"
	"path/filepath"
	"strconv"

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
	path := filepath.Join(req.BasePath, "dyd", "heap", "files", fingerprintVersionV2)
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapStems(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "stems", fingerprintVersionV2)
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapSprouts(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "sprouts", fingerprintVersionV2)
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapDerivations(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "derivations", "roots", fingerprintVersionV2)
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
	path := filepath.Join(req.BasePath, "dyd", "heap", "secrets", fingerprintVersionV2)
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

func gardenCreateShedHeapDepthFile(path string) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(shedHeapDepthDefault)), 0o644)
}

func gardenCreateShedHeapDepthPath(basePath string, segments ...string) string {
	parts := append([]string{basePath, "dyd", "shed", "heap"}, segments...)
	return filepath.Join(parts...)
}

func gardenCreateShedHeapFilesDepth(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	err := gardenCreateShedHeapDepthFile(gardenCreateShedHeapDepthPath(req.BasePath, "files", "depth"))
	return err, req
}

func gardenCreateShedHeapSecretsDepth(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	err := gardenCreateShedHeapDepthFile(gardenCreateShedHeapDepthPath(req.BasePath, "secrets", "depth"))
	return err, req
}

func gardenCreateShedHeapStemsDepth(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	err := gardenCreateShedHeapDepthFile(gardenCreateShedHeapDepthPath(req.BasePath, "stems", "depth"))
	return err, req
}

func gardenCreateShedHeapSproutsDepth(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	err := gardenCreateShedHeapDepthFile(gardenCreateShedHeapDepthPath(req.BasePath, "sprouts", "depth"))
	return err, req
}

func gardenCreateShedHeapDerivationsRootsDepth(
	ctx *task.ExecutionContext,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	err := gardenCreateShedHeapDepthFile(gardenCreateShedHeapDepthPath(req.BasePath, "derivations", "roots", "depth"))
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
			task.Parallel6(
				gardenCreateHeapFiles,
				gardenCreateHeapStems,
				gardenCreateHeapSprouts,
				gardenCreateHeapDerivations,
				gardenCreateHeapContexts,
				gardenCreateHeapSecrets,
			),
			func(
				ctx *task.ExecutionContext,
				res task.Tuple6[gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest],
			) (error, gardenCreateRequest) {
				return nil, res.A
			},
		),
		task.Series3(
			gardenCreateShed,
			task.Parallel6(
				gardenCreateShedScopes,
				gardenCreateShedHeapFilesDepth,
				gardenCreateShedHeapSecretsDepth,
				gardenCreateShedHeapStemsDepth,
				gardenCreateShedHeapSproutsDepth,
				gardenCreateShedHeapDerivationsRootsDepth,
			),
			func(
				ctx *task.ExecutionContext,
				res task.Tuple6[gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest],
			) (error, gardenCreateRequest) {
				return nil, res.A
			},
		),
		gardenCreateRoots,
		gardenCreateSprouts,
		gardenCreateTypeFile,
	),
	func(
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

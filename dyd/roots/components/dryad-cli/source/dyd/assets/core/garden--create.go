package core

import (
	"context"
	"dryad/internal/filepath"
	"dryad/internal/os"
	"fmt"
	"strconv"

	task "dryad/task"
)

type gardenCreateRequest struct {
	BasePath string
}

func gardenPrepareRequest(
	ctx context.Context,
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
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeap(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapFiles(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "files", fingerprintVersionV2)
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapStems(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "stems", fingerprintVersionV2)
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapSprouts(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "sprouts", fingerprintVersionV2)
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapDerivations(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "derivations", "roots", fingerprintVersionV2)
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapContexts(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "contexts")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapSecrets(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "secrets", fingerprintVersionV2)
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateShed(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "shed")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateShedScopes(
	ctx context.Context,
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
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	err := gardenCreateShedHeapDepthFile(gardenCreateShedHeapDepthPath(req.BasePath, "files", "depth"))
	return err, req
}

func gardenCreateShedHeapSecretsDepth(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	err := gardenCreateShedHeapDepthFile(gardenCreateShedHeapDepthPath(req.BasePath, "secrets", "depth"))
	return err, req
}

func gardenCreateShedHeapStemsDepth(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	err := gardenCreateShedHeapDepthFile(gardenCreateShedHeapDepthPath(req.BasePath, "stems", "depth"))
	return err, req
}

func gardenCreateShedHeapSproutsDepth(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	err := gardenCreateShedHeapDepthFile(gardenCreateShedHeapDepthPath(req.BasePath, "sprouts", "depth"))
	return err, req
}

func gardenCreateShedHeapDerivationsRootsDepth(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	err := gardenCreateShedHeapDepthFile(gardenCreateShedHeapDepthPath(req.BasePath, "derivations", "roots", "depth"))
	return err, req
}

func gardenCreateRoots(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "roots")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateSprouts(
	ctx context.Context,
	req gardenCreateRequest,
) (error, gardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "sprouts")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateTypeFile(
	ctx context.Context,
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
				ctx context.Context,
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
				ctx context.Context,
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
		ctx context.Context,
		res task.Tuple5[gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest, gardenCreateRequest],
	) (error, gardenCreateRequest) {
		return nil, res.A
	},
)

func (ug *UnsafeGardenReference) Create(ctx context.Context) (error, *SafeGardenReference) {
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

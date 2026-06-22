package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"dryad/internal/os"
	// zlog "github.com/rs/zerolog/log"
)

func resolveHeapFilesReference(ctx *task.ExecutionContext, heapFiles *UnsafeHeapFilesReference) (error, *SafeHeapFilesReference) {
	var heapFilesExists bool
	var err error
	var safeRef SafeHeapFilesReference

	heapFilesExists, err = fileExists(heapFiles.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapFilesExists {
		// err = os.Mkdir(heapFiles.BasePath, os.ModePerm)
		err, _ := fs2.Mkdir2(
			ctx,
			fs2.MkdirRequest{
				Path: heapFiles.BasePath,
				Mode: os.ModePerm,
			},
		)
		if err != nil {
			return err, nil
		}
	}

	err, _ = fs2.Mkdir2(
		ctx,
		fs2.MkdirRequest{
			Path:      heapFilesVersionDir(heapFiles.BasePath),
			Mode:      os.ModePerm,
			Recursive: true,
		},
	)
	if err != nil {
		return err, nil
	}

	safeRef = SafeHeapFilesReference{
		BasePath: heapFiles.BasePath,
		Heap:     heapFiles.Heap,
	}

	return nil, &safeRef
}

var memoResolveHeapFilesReference = task.Memoize(
	resolveHeapFilesReference,
	func(ctx *task.ExecutionContext, heapFiles *UnsafeHeapFilesReference) (error, any) {
		type Key struct {
			Group      string
			BasePath   string
			HeapPath   string
			GardenPath string
		}

		heapPath := ""
		gardenPath := ""
		if heapFiles.Heap != nil {
			heapPath = heapFiles.Heap.BasePath
			if heapFiles.Heap.Garden != nil {
				gardenPath = heapFiles.Heap.Garden.BasePath
			}
		}

		return nil, Key{
			Group:      "HeapFiles.Resolve",
			BasePath:   heapFiles.BasePath,
			HeapPath:   heapPath,
			GardenPath: gardenPath,
		}
	},
)

func (heapFiles *UnsafeHeapFilesReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapFilesReference) {
	return memoResolveHeapFilesReference(ctx, heapFiles)
}

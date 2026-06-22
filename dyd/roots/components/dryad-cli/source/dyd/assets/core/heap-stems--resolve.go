package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"dryad/internal/os"
	// zlog "github.com/rs/zerolog/log"
)

func resolveHeapStemsReference(ctx *task.ExecutionContext, heapStems *UnsafeHeapStemsReference) (error, *SafeHeapStemsReference) {
	var heapStemsExists bool
	var err error
	var safeRef SafeHeapStemsReference

	heapStemsExists, err = fileExists(heapStems.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapStemsExists {
		// err = os.Mkdir(heapStems.BasePath, os.ModePerm)
		err, _ := fs2.Mkdir2(
			ctx,
			fs2.MkdirRequest{
				Path: heapStems.BasePath,
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
			Path:      heapStemsVersionDir(heapStems.BasePath),
			Mode:      os.ModePerm,
			Recursive: true,
		},
	)
	if err != nil {
		return err, nil
	}

	safeRef = SafeHeapStemsReference{
		BasePath: heapStems.BasePath,
		Heap:     heapStems.Heap,
	}

	return nil, &safeRef
}

var memoResolveHeapStemsReference = task.Memoize(
	resolveHeapStemsReference,
	func(ctx *task.ExecutionContext, heapStems *UnsafeHeapStemsReference) (error, any) {
		type Key struct {
			Group      string
			BasePath   string
			HeapPath   string
			GardenPath string
		}

		heapPath := ""
		gardenPath := ""
		if heapStems.Heap != nil {
			heapPath = heapStems.Heap.BasePath
			if heapStems.Heap.Garden != nil {
				gardenPath = heapStems.Heap.Garden.BasePath
			}
		}

		return nil, Key{
			Group:      "HeapStems.Resolve",
			BasePath:   heapStems.BasePath,
			HeapPath:   heapPath,
			GardenPath: gardenPath,
		}
	},
)

func (heapStems *UnsafeHeapStemsReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapStemsReference) {
	return memoResolveHeapStemsReference(ctx, heapStems)
}

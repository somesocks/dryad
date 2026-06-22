package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"dryad/internal/os"
	// zlog "github.com/rs/zerolog/log"
)

func resolveHeapSproutsReference(ctx *task.ExecutionContext, heapSprouts *UnsafeHeapSproutsReference) (error, *SafeHeapSproutsReference) {
	var heapSproutsExists bool
	var err error
	var safeRef SafeHeapSproutsReference

	heapSproutsExists, err = fileExists(heapSprouts.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapSproutsExists {
		err, _ := fs2.Mkdir2(
			ctx,
			fs2.MkdirRequest{
				Path: heapSprouts.BasePath,
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
			Path:      heapSproutsVersionDir(heapSprouts.BasePath),
			Mode:      os.ModePerm,
			Recursive: true,
		},
	)
	if err != nil {
		return err, nil
	}

	safeRef = SafeHeapSproutsReference{
		BasePath: heapSprouts.BasePath,
		Heap:     heapSprouts.Heap,
	}

	return nil, &safeRef
}

var memoResolveHeapSproutsReference = task.Memoize(
	resolveHeapSproutsReference,
	func(ctx *task.ExecutionContext, heapSprouts *UnsafeHeapSproutsReference) (error, any) {
		type Key struct {
			Group      string
			BasePath   string
			HeapPath   string
			GardenPath string
		}

		heapPath := ""
		gardenPath := ""
		if heapSprouts.Heap != nil {
			heapPath = heapSprouts.Heap.BasePath
			if heapSprouts.Heap.Garden != nil {
				gardenPath = heapSprouts.Heap.Garden.BasePath
			}
		}

		return nil, Key{
			Group:      "HeapSprouts.Resolve",
			BasePath:   heapSprouts.BasePath,
			HeapPath:   heapPath,
			GardenPath: gardenPath,
		}
	},
)

func (heapSprouts *UnsafeHeapSproutsReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapSproutsReference) {
	return memoResolveHeapSproutsReference(ctx, heapSprouts)
}

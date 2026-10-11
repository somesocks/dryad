package core

import (
	"context"
	fs2 "dryad/filesystem"
	"dryad/task"

	"dryad/internal/os"

	zlog "github.com/rs/zerolog/log"
)

func resolveHeapReference(ctx context.Context, heap *UnsafeHeapReference) (error, *SafeHeapReference) {
	zlog.Trace().
		Str("path", heap.BasePath).
		Msg("UnsafeHeapReference.Resolve")

	var heapExists bool
	var err error
	var safeRef SafeHeapReference

	heapExists, err = fileExists(heap.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapExists {
		// err := os.Mkdir(heap.BasePath, os.ModePerm)
		err, _ := fs2.Mkdir2(
			ctx,
			fs2.MkdirRequest{
				Path: heap.BasePath,
				Mode: os.ModePerm,
			},
		)
		if err != nil {
			return err, nil
		}
	}

	safeRef = SafeHeapReference{
		BasePath: heap.BasePath,
		Garden:   heap.Garden,
	}

	return nil, &safeRef
}

var memoResolveHeapReference = task.Memoize(
	resolveHeapReference,
	func(ctx context.Context, heap *UnsafeHeapReference) (error, any) {
		type Key struct {
			Group      string
			BasePath   string
			GardenPath string
		}

		gardenPath := ""
		if heap.Garden != nil {
			gardenPath = heap.Garden.BasePath
		}

		return nil, Key{
			Group:      "Heap.Resolve",
			BasePath:   heap.BasePath,
			GardenPath: gardenPath,
		}
	},
)

func (heap *UnsafeHeapReference) Resolve(ctx context.Context) (error, *SafeHeapReference) {
	return memoResolveHeapReference(ctx, heap)
}

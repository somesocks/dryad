package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"dryad/internal/os"
	// zlog "github.com/rs/zerolog/log"
)

func resolveHeapSecretsReference(ctx *task.ExecutionContext, heapSecrets *UnsafeHeapSecretsReference) (error, *SafeHeapSecretsReference) {
	var heapSecretsExists bool
	var err error
	var safeRef SafeHeapSecretsReference

	heapSecretsExists, err = fileExists(heapSecrets.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapSecretsExists {
		// err = os.Mkdir(heapSecrets.BasePath, os.ModePerm)
		err, _ := fs2.Mkdir2(
			ctx,
			fs2.MkdirRequest{
				Path: heapSecrets.BasePath,
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
			Path:      heapSecretsVersionDir(heapSecrets.BasePath),
			Mode:      os.ModePerm,
			Recursive: true,
		},
	)
	if err != nil {
		return err, nil
	}

	safeRef = SafeHeapSecretsReference{
		BasePath: heapSecrets.BasePath,
		Heap:     heapSecrets.Heap,
	}

	return nil, &safeRef
}

var memoResolveHeapSecretsReference = task.Memoize(
	resolveHeapSecretsReference,
	func(ctx *task.ExecutionContext, heapSecrets *UnsafeHeapSecretsReference) (error, any) {
		type Key struct {
			Group      string
			BasePath   string
			HeapPath   string
			GardenPath string
		}

		heapPath := ""
		gardenPath := ""
		if heapSecrets.Heap != nil {
			heapPath = heapSecrets.Heap.BasePath
			if heapSecrets.Heap.Garden != nil {
				gardenPath = heapSecrets.Heap.Garden.BasePath
			}
		}

		return nil, Key{
			Group:      "HeapSecrets.Resolve",
			BasePath:   heapSecrets.BasePath,
			HeapPath:   heapPath,
			GardenPath: gardenPath,
		}
	},
)

func (heapSecrets *UnsafeHeapSecretsReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapSecretsReference) {
	return memoResolveHeapSecretsReference(ctx, heapSecrets)
}

package core

import (
	"dryad/task"

	"dryad/internal/os"
	// zlog "github.com/rs/zerolog/log"
)

func resolveRootsReference(ctx *task.ExecutionContext, ur *UnsafeRootsReference) (error, *SafeRootsReference) {
	var rootsExists bool
	var err error
	var safeRef SafeRootsReference

	rootsExists, err = fileExists(ur.BasePath)
	if err != nil {
		return err, nil
	}

	if !rootsExists {
		err = os.Mkdir(ur.BasePath, os.ModePerm)
		if err != nil {
			return err, nil
		}
	}

	safeRef = SafeRootsReference{
		BasePath: ur.BasePath,
		Garden:   ur.Garden,
	}

	return nil, &safeRef
}

var memoResolveRootsReference = task.Memoize(
	resolveRootsReference,
	func(ctx *task.ExecutionContext, ur *UnsafeRootsReference) (error, any) {
		type Key struct {
			Group      string
			BasePath   string
			GardenPath string
		}

		gardenPath := ""
		if ur.Garden != nil {
			gardenPath = ur.Garden.BasePath
		}

		return nil, Key{
			Group:      "Roots.Resolve",
			BasePath:   ur.BasePath,
			GardenPath: gardenPath,
		}
	},
)

func (ur *UnsafeRootsReference) Resolve(ctx *task.ExecutionContext) (error, *SafeRootsReference) {
	return memoResolveRootsReference(ctx, ur)
}

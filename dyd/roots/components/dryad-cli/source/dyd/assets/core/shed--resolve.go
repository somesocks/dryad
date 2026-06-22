package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"dryad/internal/os"
)

func resolveShedReference(ctx *task.ExecutionContext, shed *UnsafeShedReference) (error, *SafeShedReference) {
	var shedExists bool
	var err error
	var safeRef SafeShedReference

	shedExists, err = fileExists(shed.BasePath)
	if err != nil {
		return err, nil
	}

	if !shedExists {
		err, _ = fs2.Mkdir2(
			ctx,
			fs2.MkdirRequest{
				Path: shed.BasePath,
				Mode: os.ModePerm,
			},
		)
		if err != nil {
			return err, nil
		}
	}

	safeRef = SafeShedReference{
		BasePath: shed.BasePath,
		Garden:   shed.Garden,
	}

	return nil, &safeRef
}

var memoResolveShedReference = task.Memoize(
	resolveShedReference,
	func(ctx *task.ExecutionContext, shed *UnsafeShedReference) (error, any) {
		type Key struct {
			Group      string
			BasePath   string
			GardenPath string
		}

		gardenPath := ""
		if shed.Garden != nil {
			gardenPath = shed.Garden.BasePath
		}

		return nil, Key{
			Group:      "Shed.Resolve",
			BasePath:   shed.BasePath,
			GardenPath: gardenPath,
		}
	},
)

func (shed *UnsafeShedReference) Resolve(ctx *task.ExecutionContext) (error, *SafeShedReference) {
	return memoResolveShedReference(ctx, shed)
}

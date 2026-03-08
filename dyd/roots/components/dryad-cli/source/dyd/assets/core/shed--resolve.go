package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"dryad/internal/os"
)

func (shed *UnsafeShedReference) Resolve(ctx *task.ExecutionContext) (error, *SafeShedReference) {
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

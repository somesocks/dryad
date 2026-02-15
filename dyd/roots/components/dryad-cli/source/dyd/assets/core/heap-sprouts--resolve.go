package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"os"
	// zlog "github.com/rs/zerolog/log"
)

func (heapSprouts *UnsafeHeapSproutsReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapSproutsReference) {
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

	safeRef = SafeHeapSproutsReference{
		BasePath: heapSprouts.BasePath,
		Heap:     heapSprouts.Heap,
	}

	return nil, &safeRef
}

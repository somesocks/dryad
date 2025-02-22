
package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"os"

	// zlog "github.com/rs/zerolog/log"
)


func (heapStems *UnsafeHeapStemsReference) Resolve(ctx * task.ExecutionContext) (error, *SafeHeapStemsReference) {
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

	safeRef = SafeHeapStemsReference{
		BasePath: heapStems.BasePath,
		Heap: heapStems.Heap,
	}

	return nil, &safeRef 
}
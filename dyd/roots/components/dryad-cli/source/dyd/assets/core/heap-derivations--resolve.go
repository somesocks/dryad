
package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"os"

	// zlog "github.com/rs/zerolog/log"
)


func (heapDerivations *UnsafeHeapDerivationsReference) Resolve(ctx * task.ExecutionContext) (error, *SafeHeapDerivationsReference) {
	var heapDerivationsExists bool
	var err error
	var safeRef SafeHeapDerivationsReference

	heapDerivationsExists, err = fileExists(heapDerivations.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapDerivationsExists {
		// err = os.Mkdir(heapDerivations.BasePath, os.ModePerm)
		err, _ := fs2.Mkdir2(
			ctx,
			fs2.MkdirRequest{
				Path: heapDerivations.BasePath,
				Mode: os.ModePerm,
			},
		)
		if err != nil {
			return err, nil
		}
	}

	safeRef = SafeHeapDerivationsReference{
		BasePath: heapDerivations.BasePath,
		Heap: heapDerivations.Heap,
	}

	return nil, &safeRef 
}
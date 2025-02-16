
package core

import (
	"dryad/task"

	"os"

	// zlog "github.com/rs/zerolog/log"
)


func (heap *UnsafeHeapReference) Resolve(ctx * task.ExecutionContext) (error, *SafeHeapReference) {
	var heapExists bool
	var err error
	var safeRef SafeHeapReference

	heapExists, err = fileExists(heap.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapExists {
		err = os.Mkdir(heap.BasePath, os.ModePerm)
		if err != nil {
			return err, nil
		}
	}

	safeRef = SafeHeapReference{
		BasePath: heap.BasePath,
		Garden: heap.Garden,
	}

	return nil, &safeRef 
}
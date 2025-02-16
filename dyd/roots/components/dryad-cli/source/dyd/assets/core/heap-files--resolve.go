
package core

import (
	"dryad/task"

	"os"

	// zlog "github.com/rs/zerolog/log"
)


func (heapFiles *UnsafeHeapFilesReference) Resolve(ctx * task.ExecutionContext) (error, *SafeHeapFilesReference) {
	var heapFilesExists bool
	var err error
	var safeRef SafeHeapFilesReference

	heapFilesExists, err = fileExists(heapFiles.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapFilesExists {
		err = os.Mkdir(heapFiles.BasePath, os.ModePerm)
		if err != nil {
			return err, nil
		}
	}

	safeRef = SafeHeapFilesReference{
		BasePath: heapFiles.BasePath,
		Heap: heapFiles.Heap,
	}

	return nil, &safeRef 
}

package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"os"

	// zlog "github.com/rs/zerolog/log"
)


func (heapSecrets *UnsafeHeapSecretsReference) Resolve(ctx * task.ExecutionContext) (error, *SafeHeapSecretsReference) {
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

	safeRef = SafeHeapSecretsReference{
		BasePath: heapSecrets.BasePath,
		Heap: heapSecrets.Heap,
	}

	return nil, &safeRef 
}
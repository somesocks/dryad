
package core

import (
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
		err = os.Mkdir(heapSecrets.BasePath, os.ModePerm)
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
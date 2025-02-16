
package core

import (
	"dryad/task"
	"errors"

	"path/filepath"

	// "os"

	// zlog "github.com/rs/zerolog/log"
)


func (heapSecret *UnsafeHeapSecretReference) Resolve(ctx * task.ExecutionContext) (error, *SafeHeapSecretReference) {
	var heapSecretExists bool
	var err error
	var safeRef SafeHeapSecretReference

	heapSecretExists, err = fileExists(heapSecret.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapSecretExists {
		return errors.New("unable to resolve stem"), nil
	}

	safeRef = SafeHeapSecretReference{
		BasePath: heapSecret.BasePath,
		// TODO: should this be read from fingerprint file?
		Fingerprint: filepath.Base(heapSecret.BasePath),
		Secrets: heapSecret.Secrets,
	}

	return nil, &safeRef 
}

package core

import (
	"dryad/task"
	"errors"

	// "os"

	// zlog "github.com/rs/zerolog/log"
)


func (heapStem *UnsafeHeapStemReference) Resolve(ctx * task.ExecutionContext) (error, *SafeHeapStemReference) {
	var heapStemExists bool
	var err error
	var safeRef SafeHeapStemReference

	heapStemExists, err = fileExists(heapStem.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapStemExists {
		return errors.New("unable to resolve stem"), nil
	}

	safeRef = SafeHeapStemReference{
		BasePath: heapStem.BasePath,
		Stems: heapStem.Stems,
	}

	return nil, &safeRef 
}
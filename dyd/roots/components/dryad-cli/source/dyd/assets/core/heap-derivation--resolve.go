
package core

import (
	"dryad/task"
	"errors"

	"path/filepath"

	// "os"

	// zlog "github.com/rs/zerolog/log"
)


func (heapDerivation *UnsafeHeapDerivationReference) Resolve(ctx * task.ExecutionContext) (error, *SafeHeapDerivationReference) {
	var heapDerivationExists bool
	var err error
	var safeRef SafeHeapDerivationReference

	heapDerivationExists, err = fileExists(heapDerivation.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapDerivationExists {
		return errors.New("unable to resolve derivation"), nil
	}

	err, heapStems := heapDerivation.Derivations.Heap.Stems().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	sourceStem := heapStems.Stem(filepath.Base(heapDerivation.BasePath))

	resultPath, err := filepath.EvalSymlinks(heapDerivation.BasePath)
	if err != nil {
		return err, nil
	}

	resultStem := heapStems.Stem(filepath.Base(resultPath))

	safeRef = SafeHeapDerivationReference{
		BasePath: heapDerivation.BasePath,
		Source: sourceStem,
		Result: resultStem,
		Derivations: heapDerivation.Derivations,
	}

	return nil, &safeRef 
}
package core

import (
	dydfs "dryad/filesystem"
	"path/filepath"
	"dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (derivations *SafeHeapDerivationsReference) Add(
	ctx *task.ExecutionContext,
	sourceFingerprint string,
	resultFingerprint string,
) (error, *SafeHeapDerivationReference) {

	derivationPath := filepath.Join(derivations.BasePath, sourceFingerprint)	

	derivationTarget , err := filepath.Rel(
		derivations.BasePath,
		filepath.Join(derivations.Heap.BasePath, "stems", resultFingerprint),
	)
	if err != nil {
		return err, nil
	}

	err, _ = dydfs.Symlink(
		ctx,
		dydfs.SymlinkRequest{
			Path: derivationPath,
			Target: derivationTarget,
		},
	)
	if err != nil {
		return err, nil
	}

	err, heapStems := derivations.Heap.Stems().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	sourceStem := heapStems.Stem(sourceFingerprint)

	resultStem := heapStems.Stem(resultFingerprint)

	safeRef := SafeHeapDerivationReference{
		BasePath: derivationPath,
		Source: sourceStem,
		Result: resultStem,
		Derivations: derivations,
	}

	return nil, &safeRef 

}
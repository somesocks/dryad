package core

import (
	"dryad/internal/os"
	"dryad/task"
	"errors"
	"io/fs"
	"path/filepath"
	// zlog "github.com/rs/zerolog/log"
)

func (derivations *SafeHeapDerivationsReference) Add(
	ctx *task.ExecutionContext,
	sourceFingerprint string,
	resultFingerprint string,
) (error, *SafeHeapDerivationReference) {

	err, derivationPath := heapDerivationsRootsFingerprintPath(ctx, derivations.Heap.Garden, derivations.BasePath, sourceFingerprint)
	if err != nil {
		return err, nil
	}
	derivationsRootsPath := filepath.Dir(derivationPath)

	tempFile, err := os.CreateTemp(
		derivationsRootsPath,
		".tmp-"+sourceFingerprint+"-*",
	)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(derivationsRootsPath, os.ModePerm)
		if err != nil {
			return err, nil
		}
		tempFile, err = os.CreateTemp(
			derivationsRootsPath,
			".tmp-"+sourceFingerprint+"-*",
		)
	}
	if err != nil {
		return err, nil
	}
	tempPath := tempFile.Name()
	// Best effort cleanup. Crash/power-loss can still leave tmp files behind.
	defer os.Remove(tempPath)

	_, err = tempFile.WriteString(resultFingerprint)
	if err != nil {
		return err, nil
	}

	err = tempFile.Close()
	if err != nil {
		return err, nil
	}

	// Publish atomically without overwriting an existing derivation entry.
	err = os.Link(tempPath, derivationPath)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return err, nil
		}
	}

	err, heapStems := derivations.Heap.Stems().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	err, sourceStemPath := heapStemsFingerprintPath(ctx, heapStems.Heap.Garden, heapStems.BasePath, sourceFingerprint)
	if err != nil {
		return err, nil
	}
	err, resultStemPath := heapStemsFingerprintPath(ctx, heapStems.Heap.Garden, heapStems.BasePath, resultFingerprint)
	if err != nil {
		return err, nil
	}

	sourceStem := heapStems.Stem(sourceFingerprint)
	sourceStem.BasePath = sourceStemPath
	resultStem := heapStems.Stem(resultFingerprint)
	resultStem.BasePath = resultStemPath

	safeRef := SafeHeapDerivationReference{
		BasePath:          derivationPath,
		SourceFingerprint: sourceFingerprint,
		ResultFingerprint: resultFingerprint,
		Source:            sourceStem,
		Result:            resultStem,
		Derivations:       derivations,
	}

	return nil, &safeRef

}

package core

import (
	"dryad/internal/filepath"
	"dryad/task"
	"errors"
	// "os"
	// zlog "github.com/rs/zerolog/log"
)

func (heapStem *UnsafeHeapStemReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapStemReference) {
	var heapStemExists bool
	var err error
	var safeRef SafeHeapStemReference
	var resolvedPath string

	err, resolvedPath = heapStemsFingerprintPath(ctx, heapStem.Stems.Heap.Garden, heapStem.Stems.BasePath, heapStem.Fingerprint)
	if err != nil {
		return err, nil
	}
	if heapStem.BasePath != "" && filepath.Clean(heapStem.BasePath) != filepath.Clean(resolvedPath) {
		return errors.New("unable to resolve stem"), nil
	}

	heapStemExists, err = fileExists(resolvedPath)
	if err != nil {
		return err, nil
	}

	if !heapStemExists {
		return errors.New("unable to resolve stem"), nil
	}

	safeRef = SafeHeapStemReference{
		BasePath:    resolvedPath,
		Fingerprint: heapStem.Fingerprint,
		Stems:       heapStem.Stems,
	}

	return nil, &safeRef
}

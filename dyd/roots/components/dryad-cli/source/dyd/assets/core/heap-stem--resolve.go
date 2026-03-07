package core

import (
	"dryad/task"
	"errors"
	"path/filepath"
	// "os"
	// zlog "github.com/rs/zerolog/log"
)

func (heapStem *UnsafeHeapStemReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapStemReference) {
	var heapStemExists bool
	var err error
	var safeRef SafeHeapStemReference
	var expectedPath string

	expectedPath, err = heapStemsFingerprintPath(heapStem.Stems.BasePath, heapStem.Fingerprint)
	if err != nil {
		return err, nil
	}
	if filepath.Clean(heapStem.BasePath) != filepath.Clean(expectedPath) {
		return errors.New("unable to resolve stem"), nil
	}

	heapStemExists, err = fileExists(heapStem.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapStemExists {
		return errors.New("unable to resolve stem"), nil
	}

	safeRef = SafeHeapStemReference{
		BasePath:    heapStem.BasePath,
		Fingerprint: heapStem.Fingerprint,
		Stems:       heapStem.Stems,
	}

	return nil, &safeRef
}

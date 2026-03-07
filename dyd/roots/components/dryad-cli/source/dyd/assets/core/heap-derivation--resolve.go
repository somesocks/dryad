package core

import (
	"dryad/internal/os"
	"dryad/task"
	"errors"
	"io/fs"

	"path/filepath"
	"strings"
	// "os"
	// zlog "github.com/rs/zerolog/log"
)

var ErrUnresolvableHeapDerivation = errors.New("unable to resolve derivation")

func (heapDerivation *UnsafeHeapDerivationReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapDerivationReference) {
	var err error
	var safeRef SafeHeapDerivationReference
	var expectedPath string

	expectedPath, err = heapDerivationsRootsFingerprintPath(
		heapDerivation.Derivations.BasePath,
		heapDerivation.SourceFingerprint,
	)
	if err != nil {
		return err, nil
	}
	if filepath.Clean(heapDerivation.BasePath) != filepath.Clean(expectedPath) {
		return ErrUnresolvableHeapDerivation, nil
	}

	info, err := os.Lstat(heapDerivation.BasePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrUnresolvableHeapDerivation, nil
		}
		return err, nil
	}
	if !info.Mode().IsRegular() {
		return ErrUnresolvableHeapDerivation, nil
	}

	heapStems := &SafeHeapStemsReference{
		BasePath: filepath.Join(heapDerivation.Derivations.Heap.BasePath, "stems"),
		Heap:     heapDerivation.Derivations.Heap,
	}

	sourceStem := heapStems.Stem(heapDerivation.SourceFingerprint)

	resultFingerprintBytes, err := os.ReadFile(heapDerivation.BasePath)
	if err != nil {
		return err, nil
	}
	resultFingerprint := strings.TrimSpace(string(resultFingerprintBytes))
	if resultFingerprint == "" {
		return ErrUnresolvableHeapDerivation, nil
	}

	resultStemPath, err := heapStemsFingerprintPath(heapStems.BasePath, resultFingerprint)
	if err != nil {
		return err, nil
	}
	_, err = os.Stat(resultStemPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrUnresolvableHeapDerivation, nil
		}
		return err, nil
	}

	resultStem := heapStems.Stem(resultFingerprint)

	safeRef = SafeHeapDerivationReference{
		BasePath:          heapDerivation.BasePath,
		SourceFingerprint: heapDerivation.SourceFingerprint,
		ResultFingerprint: resultFingerprint,
		Source:            sourceStem,
		Result:            resultStem,
		Derivations:       heapDerivation.Derivations,
	}

	return nil, &safeRef
}

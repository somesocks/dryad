package core

import (
	"dryad/internal/os"
	"dryad/task"
	"errors"
	"io/fs"

	"dryad/internal/filepath"
	"strings"
	// "os"
	// zlog "github.com/rs/zerolog/log"
)

var ErrUnresolvableHeapDerivation = errors.New("unable to resolve derivation")

func (heapDerivation *UnsafeHeapDerivationReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapDerivationReference) {
	var err error
	var safeRef SafeHeapDerivationReference
	var derivationPath string

	err, derivationPath = heapDerivationsRootsFingerprintPath(
		ctx,
		heapDerivation.Derivations.Heap.Garden,
		heapDerivation.Derivations.BasePath,
		heapDerivation.SourceFingerprint,
	)
	if err != nil {
		return err, nil
	}
	if heapDerivation.BasePath != "" && filepath.Clean(heapDerivation.BasePath) != filepath.Clean(derivationPath) {
		return ErrUnresolvableHeapDerivation, nil
	}

	info, err := os.Lstat(derivationPath)
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

	err, sourceStemPath := heapStemsFingerprintPath(ctx, heapStems.Heap.Garden, heapStems.BasePath, heapDerivation.SourceFingerprint)
	if err != nil {
		return err, nil
	}
	sourceStem := heapStems.Stem(heapDerivation.SourceFingerprint)
	sourceStem.BasePath = sourceStemPath

	resultFingerprintBytes, err := os.ReadFile(derivationPath)
	if err != nil {
		return err, nil
	}
	resultFingerprint := strings.TrimSpace(string(resultFingerprintBytes))
	if resultFingerprint == "" {
		return ErrUnresolvableHeapDerivation, nil
	}

	err, resultStemPath := heapStemsFingerprintPath(ctx, heapStems.Heap.Garden, heapStems.BasePath, resultFingerprint)
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
	resultStem.BasePath = resultStemPath

	safeRef = SafeHeapDerivationReference{
		BasePath:          derivationPath,
		SourceFingerprint: heapDerivation.SourceFingerprint,
		ResultFingerprint: resultFingerprint,
		Source:            sourceStem,
		Result:            resultStem,
		Derivations:       heapDerivation.Derivations,
	}

	return nil, &safeRef
}

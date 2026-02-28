package core

import (
	"dryad/task"
	"errors"
	"io/fs"
	"os"

	"path/filepath"
	"strings"
	// "os"
	// zlog "github.com/rs/zerolog/log"
)

func (heapDerivation *UnsafeHeapDerivationReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapDerivationReference) {
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

	info, err := os.Lstat(heapDerivation.BasePath)
	if err != nil {
		return err, nil
	}
	if !info.Mode().IsRegular() {
		return errors.New("unable to resolve derivation"), nil
	}

	err, heapStems := heapDerivation.Derivations.Heap.Stems().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	sourceStem := heapStems.Stem(filepath.Base(heapDerivation.BasePath))

	resultFingerprintBytes, err := os.ReadFile(heapDerivation.BasePath)
	if err != nil {
		return err, nil
	}
	resultFingerprint := strings.TrimSpace(string(resultFingerprintBytes))
	if resultFingerprint == "" {
		return errors.New("unable to resolve derivation"), nil
	}

	resultStemPath := filepath.Join(heapDerivation.Derivations.Heap.BasePath, "stems", resultFingerprint)
	_, err = os.Stat(resultStemPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return errors.New("unable to resolve derivation"), nil
		}
		return err, nil
	}

	resultStem := heapStems.Stem(resultFingerprint)

	safeRef = SafeHeapDerivationReference{
		BasePath:    heapDerivation.BasePath,
		Source:      sourceStem,
		Result:      resultStem,
		Derivations: heapDerivation.Derivations,
	}

	return nil, &safeRef
}

package core

import (
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"errors"
	"io/fs"
	// "os"
	// zlog "github.com/rs/zerolog/log"
)

func (heapDerivation *UnsafeHeapDerivationReference) Exists(ctx *task.ExecutionContext) (error, bool) {
	err, derivationPath := heapDerivationsRootsFingerprintPath(
		ctx,
		heapDerivation.Derivations.Heap.Garden,
		heapDerivation.Derivations.BasePath,
		heapDerivation.SourceFingerprint,
	)
	if err != nil {
		return err, false
	}
	if heapDerivation.BasePath != "" && filepath.Clean(heapDerivation.BasePath) != filepath.Clean(derivationPath) {
		return nil, false
	}

	info, err := os.Lstat(derivationPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false
		}
		return err, false
	}

	return nil, info.Mode().IsRegular()
}

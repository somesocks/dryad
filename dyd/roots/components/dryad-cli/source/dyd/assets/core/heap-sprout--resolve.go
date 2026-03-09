package core

import (
	"dryad/task"
	"errors"
	"path/filepath"
	// "os"
	// zlog "github.com/rs/zerolog/log"
)

func (heapSprout *UnsafeHeapSproutReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapSproutReference) {
	var heapSproutExists bool
	var err error
	var safeRef SafeHeapSproutReference
	var resolvedPath string

	err, resolvedPath = heapSproutsFingerprintPath(ctx, heapSprout.Sprouts.Heap.Garden, heapSprout.Sprouts.BasePath, heapSprout.Fingerprint)
	if err != nil {
		return err, nil
	}
	if heapSprout.BasePath != "" && filepath.Clean(heapSprout.BasePath) != filepath.Clean(resolvedPath) {
		return errors.New("unable to resolve sprout"), nil
	}

	heapSproutExists, err = fileExists(resolvedPath)
	if err != nil {
		return err, nil
	}

	if !heapSproutExists {
		return errors.New("unable to resolve sprout"), nil
	}

	safeRef = SafeHeapSproutReference{
		BasePath:    resolvedPath,
		Fingerprint: heapSprout.Fingerprint,
		Sprouts:     heapSprout.Sprouts,
	}

	return nil, &safeRef
}

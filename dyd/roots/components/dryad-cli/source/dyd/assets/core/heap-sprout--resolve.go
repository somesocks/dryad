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
	var expectedPath string

	expectedPath, err = heapSproutsFingerprintPath(heapSprout.Sprouts.BasePath, heapSprout.Fingerprint)
	if err != nil {
		return err, nil
	}
	if filepath.Clean(heapSprout.BasePath) != filepath.Clean(expectedPath) {
		return errors.New("unable to resolve sprout"), nil
	}

	heapSproutExists, err = fileExists(heapSprout.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapSproutExists {
		return errors.New("unable to resolve sprout"), nil
	}

	safeRef = SafeHeapSproutReference{
		BasePath:    heapSprout.BasePath,
		Fingerprint: heapSprout.Fingerprint,
		Sprouts:     heapSprout.Sprouts,
	}

	return nil, &safeRef
}

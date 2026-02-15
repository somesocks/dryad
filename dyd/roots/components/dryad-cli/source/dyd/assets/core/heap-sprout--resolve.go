package core

import (
	"dryad/task"
	"errors"
	// "os"
	// zlog "github.com/rs/zerolog/log"
)

func (heapSprout *UnsafeHeapSproutReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapSproutReference) {
	var heapSproutExists bool
	var err error
	var safeRef SafeHeapSproutReference

	heapSproutExists, err = fileExists(heapSprout.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapSproutExists {
		return errors.New("unable to resolve sprout"), nil
	}

	safeRef = SafeHeapSproutReference{
		BasePath: heapSprout.BasePath,
		Sprouts:  heapSprout.Sprouts,
	}

	return nil, &safeRef
}

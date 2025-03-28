
package core

import (
	// fs2 "dryad/filesystem"
	"dryad/task"

	// zlog "github.com/rs/zerolog/log"
)

func (rootTraits *UnsafeRootTraitsReference) Resolve(ctx * task.ExecutionContext) (error, *SafeRootTraitsReference) {
	var rootTraitsExists bool
	var err error
	var safeRef SafeRootTraitsReference

	rootTraitsExists, err = fileExists(rootTraits.BasePath)
	if err != nil {
		return err, nil
	}

	if !rootTraitsExists {
		return nil, nil
	}

	safeRef = SafeRootTraitsReference{
		BasePath: rootTraits.BasePath,
		Root: rootTraits.Root,
	}

	return nil, &safeRef 
}
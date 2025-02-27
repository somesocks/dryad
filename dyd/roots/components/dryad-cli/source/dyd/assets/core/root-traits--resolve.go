
package core

import (
	// fs2 "dryad/filesystem"
	"dryad/task"

	// "os"

	"errors"
	// zlog "github.com/rs/zerolog/log"
)

var ErrorNoTraits = errors.New("root does not have traits")

func (rootTraits *UnsafeRootTraitsReference) Resolve(ctx * task.ExecutionContext) (error, *SafeRootTraitsReference) {
	var rootTraitsExists bool
	var err error
	var safeRef SafeRootTraitsReference

	rootTraitsExists, err = fileExists(rootTraits.BasePath)
	if err != nil {
		return err, nil
	}

	if !rootTraitsExists {
		return ErrorNoTraits, nil
	}

	safeRef = SafeRootTraitsReference{
		BasePath: rootTraits.BasePath,
		Root: rootTraits.Root,
	}

	return nil, &safeRef 
}
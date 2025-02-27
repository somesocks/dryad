
package core

import (
	"dryad/task"
	"errors"

	// "path/filepath"

	// "os"

	// zlog "github.com/rs/zerolog/log"
)

var ErrorNoTrait = errors.New("root does not have trait")

func (rootTrait *UnsafeRootTraitReference) Resolve(ctx * task.ExecutionContext) (error, *SafeRootTraitReference) {
	var rootTraitExists bool
	var err error
	var safeRef SafeRootTraitReference

	rootTraitExists, err = fileExists(rootTrait.BasePath)
	if err != nil {
		return err, nil
	}

	if !rootTraitExists {
		return ErrorNoTrait, nil
	}

	safeRef = SafeRootTraitReference{
		BasePath: rootTrait.BasePath,
		Traits: rootTrait.Traits,
	}

	return nil, &safeRef 
}
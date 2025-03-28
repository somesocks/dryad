
package core

import (
	"dryad/task"

	// zlog "github.com/rs/zerolog/log"
)

func (rootTrait *UnsafeRootTraitReference) Resolve(ctx * task.ExecutionContext) (error, *SafeRootTraitReference) {
	var rootTraitExists bool
	var err error
	var safeRef SafeRootTraitReference

	rootTraitExists, err = fileExists(rootTrait.BasePath)
	if err != nil {
		return err, nil
	}

	if !rootTraitExists {
		return nil, nil
	}

	safeRef = SafeRootTraitReference{
		BasePath: rootTrait.BasePath,
		Traits: rootTrait.Traits,
	}

	return nil, &safeRef 
}
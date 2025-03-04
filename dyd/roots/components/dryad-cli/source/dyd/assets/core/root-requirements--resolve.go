
package core

import (
	// fs2 "dryad/filesystem"
	"dryad/task"

	// "os"

	// "errors"
	// zlog "github.com/rs/zerolog/log"
)

func (rootRequirements *UnsafeRootRequirementsReference) Resolve(ctx * task.ExecutionContext) (error, *SafeRootRequirementsReference) {
	var rootRequirementsExists bool
	var err error
	var safeRef SafeRootRequirementsReference

	rootRequirementsExists, err = fileExists(rootRequirements.BasePath)
	if err != nil {
		return err, nil
	}

	if !rootRequirementsExists {
		return nil, nil
	}

	safeRef = SafeRootRequirementsReference{
		BasePath: rootRequirements.BasePath,
		Root: rootRequirements.Root,
	}

	return nil, &safeRef 
}
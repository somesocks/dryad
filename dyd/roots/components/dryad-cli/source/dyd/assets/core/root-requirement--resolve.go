
package core

import (
	// fs2 "dryad/filesystem"
	"dryad/task"

	// "os"

	// "errors"
	// zlog "github.com/rs/zerolog/log"
)

func (rootRequirement *UnsafeRootRequirementReference) Resolve(ctx * task.ExecutionContext) (error, *SafeRootRequirementReference) {
	var rootRequirementExists bool
	var err error
	var safeRef SafeRootRequirementReference

	rootRequirementExists, err = fileExists(rootRequirement.BasePath)
	if err != nil {
		return err, nil
	}

	if !rootRequirementExists {
		return nil, nil
	}

	safeRef = SafeRootRequirementReference{
		BasePath: rootRequirement.BasePath,
		Requirements: rootRequirement.Requirements,
	}

	return nil, &safeRef 
}

package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"os"

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
		err, _ := fs2.Mkdir2(
			ctx,
			fs2.MkdirRequest{
				Path: rootRequirements.BasePath,
				Mode: os.ModePerm,
			},
		)
		if err != nil {
			return err, nil
		}
	}

	safeRef = SafeRootRequirementsReference{
		BasePath: rootRequirements.BasePath,
		Root: rootRequirements.Root,
	}

	return nil, &safeRef 
}
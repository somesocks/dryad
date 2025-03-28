
package core

import (
	"dryad/task"

	"os"

	// zlog "github.com/rs/zerolog/log"
)


func (ur *UnsafeRootsReference) Resolve(ctx * task.ExecutionContext) (error, *SafeRootsReference) {
	var rootsExists bool
	var err error
	var safeRef SafeRootsReference

	rootsExists, err = fileExists(ur.BasePath)
	if err != nil {
		return err, nil
	}

	if !rootsExists {
		err = os.Mkdir(ur.BasePath, os.ModePerm)
		if err != nil {
			return err, nil
		}
	}

	safeRef = SafeRootsReference{
		BasePath: ur.BasePath,
		Garden: ur.Garden,
	}

	return nil, &safeRef 
}
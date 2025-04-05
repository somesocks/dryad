
package core

import (
	// fs2 "dryad/filesystem"
	"dryad/task"

	"os"
	// "path/filepath"

	// "errors"
	// zlog "github.com/rs/zerolog/log"
)

func (rootRequirement *SafeRootRequirementReference) Remove(ctx * task.ExecutionContext) (error) {
	var err error

	err = os.Remove(rootRequirement.BasePath)
	return err
}
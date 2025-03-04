
package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	// "os"
	"path/filepath"

	// "errors"
	// zlog "github.com/rs/zerolog/log"
)

func (rootRequirement *SafeRootRequirementReference) Replace(ctx * task.ExecutionContext, target *SafeRootReference) (error) {
	var err error
	var linkTarget string

	linkTarget, err = filepath.Rel(
		filepath.Dir(rootRequirement.BasePath),
		target.BasePath)
	if err != nil {
		return err
	}
	
	err, _ = fs2.Symlink(
		ctx,
		fs2.SymlinkRequest{
			Path: rootRequirement.BasePath,
			Target: linkTarget,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
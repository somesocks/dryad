
package core

import (
	// fs2 "dryad/filesystem"
	"dryad/task"

	"os"
	"path/filepath"

	// "errors"
	// zlog "github.com/rs/zerolog/log"
)

func (rootRequirement *SafeRootRequirementReference) Target(ctx * task.ExecutionContext) (error, *SafeRootReference) {
	var err error
	var safeRef SafeRootReference


	linkPath, err := os.Readlink(rootRequirement.BasePath)
	if err != nil {
		return err, nil
	}

	// convert relative links to an absolute path
	if !filepath.IsAbs(linkPath) {
		linkPath = filepath.Join(
			filepath.Dir(rootRequirement.BasePath),
			linkPath,
		)
	}

	err, safeRef = rootRequirement.Requirements.Root.Roots.Root(linkPath).Resolve(ctx)
	if err != nil {
		return err, nil
	}

	return nil, &safeRef 
}
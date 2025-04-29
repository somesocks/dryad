
package core

import (
	// fs2 "dryad/filesystem"
	"dryad/task"

	// "os"
	"path/filepath"

	// "errors"
	// zlog "github.com/rs/zerolog/log"
)

type RootRequirementCopyRequest struct {
	DestRequirements *SafeRootRequirementsReference
	Unpin bool
}

func (rootRequirement *SafeRootRequirementReference) Copy(
	ctx * task.ExecutionContext,
	req RootRequirementCopyRequest,	
) (error, *SafeRootRequirementReference) {
	err, target := rootRequirement.Target(ctx)
	if err != nil {
		return err, nil
	}

	if req.Unpin {
		targetPath := target.BasePath
		relTargetPath, err := filepath.Rel(
			rootRequirement.Requirements.BasePath,
			targetPath,
		)
		if err != nil {
			return err, nil
		}
	
		newTargetPath := filepath.Join(req.DestRequirements.BasePath, relTargetPath)
		newTargetExists, err := fileExists(newTargetPath)
		if err != nil {
			return err, nil
		}

		if newTargetExists {
			err, newTarget := rootRequirement.Requirements.Root.Roots.Root(newTargetPath).Resolve(ctx)
			if err != nil {
				return err, nil
			}

			target = &newTarget
		}

	}

	alias := filepath.Base(rootRequirement.BasePath)

	err, destRequirement := req.DestRequirements.Add(
		ctx,
		RootRequirementsAddRequest{
			Dependency: target,
			Alias: alias,
		},
	)

	return err, destRequirement
}
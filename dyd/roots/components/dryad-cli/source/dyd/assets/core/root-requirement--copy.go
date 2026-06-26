package core

import (
	// fs2 "dryad/filesystem"
	"dryad/internal/os"
	"dryad/task"

	// "os"
	"dryad/internal/filepath"
	// "errors"
	// zlog "github.com/rs/zerolog/log"
)

type RootRequirementCopyRequest struct {
	DestRequirements *SafeRootRequirementsReference
	Unpin            bool
}

func (rootRequirement *SafeRootRequirementReference) Copy(
	ctx *task.ExecutionContext,
	req RootRequirementCopyRequest,
) (error, *SafeRootRequirementReference) {
	err, targetSpec := rootRequirement.TargetSpec(ctx)
	if err != nil {
		return err, nil
	}
	if rootRequirementTargetKind(targetSpec.Kind) == RootRequirementTargetKindFile {
		alias := filepath.Base(rootRequirement.BasePath)
		targetPath := targetSpec.FileSourcePath

		if req.Unpin {
			relTargetPath, err := filepath.Rel(
				rootRequirement.Requirements.BasePath,
				targetSpec.FileSourcePath,
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
				targetPath = newTargetPath
			}
		}

		relTargetPath, err := filepath.Rel(req.DestRequirements.BasePath, targetPath)
		if err != nil {
			return err, nil
		}

		return req.DestRequirements.AddFile(ctx, RootRequirementsAddFileRequest{
			Alias:  alias,
			Target: rootRequirementFileTargetString(relTargetPath, targetSpec.FileDestinationAs, targetSpec.FileDestinationInto, targetSpec.FileUnpack, targetSpec.FileFingerprint),
		})
	}

	if rootRequirementTargetKind(targetSpec.Kind) != RootRequirementTargetKindRoot {
		alias := filepath.Base(rootRequirement.BasePath)
		destRequirementPath := filepath.Join(req.DestRequirements.BasePath, alias)
		if err := os.MkdirAll(req.DestRequirements.BasePath, os.ModePerm); err != nil {
			return err, nil
		}

		contents, err := os.ReadFile(rootRequirement.BasePath)
		if err != nil {
			return err, nil
		}
		if err := os.WriteFile(destRequirementPath, contents, 0644); err != nil {
			return err, nil
		}

		return nil, &SafeRootRequirementReference{
			BasePath:     destRequirementPath,
			Requirements: req.DestRequirements,
		}
	}

	target := targetSpec.Root

	err, targetVariantSelector := variantDescriptorEncodeURL(targetSpec.VariantSelector)
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
			Dependency:                target,
			Alias:                     alias,
			DependencyVariantSelector: targetVariantSelector,
		},
	)

	return err, destRequirement
}

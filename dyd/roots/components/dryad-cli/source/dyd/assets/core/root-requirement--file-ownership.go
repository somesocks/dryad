package core

import (
	dydfs "dryad/filesystem"
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
)

type RootFileRequirementOwner struct {
	RootPath string
	Variant  VariantDescriptor
}

func rootRequirementFilePathIgnored(ctx *task.ExecutionContext, sourcePath string, path string) (error, bool) {
	info, err := os.Lstat(path)
	isDir := false
	if err == nil {
		isDir = info.IsDir()
	} else if !os.IsNotExist(err) {
		return err, false
	}
	parentDir := filepath.Dir(path)
	err, matcher := readDydIgnore(ctx, DydIgnoreRequest{
		BasePath: sourcePath,
		Path:     parentDir,
	})
	if err != nil {
		return err, false
	}
	return matcher.Match(dydfs.NewGlobPath(path, isDir))
}

func rootRequirementFileDirectoryOwnsPath(ctx *task.ExecutionContext, sourcePath string, changedPath string) (error, bool) {
	err, isWithin := rootRequirementFileIsWithin(sourcePath, changedPath)
	if err != nil {
		return err, false
	}
	if isWithin {
		if changedPath == sourcePath {
			return nil, true
		}
		err, ignored := rootRequirementFilePathIgnored(ctx, sourcePath, changedPath)
		if err != nil || ignored {
			return err, false
		}
		return nil, true
	}

	return nil, false
}

func rootRequirementFileOwnsPath(ctx *task.ExecutionContext, targetSpec *RootRequirementTargetSpec, changedPath string) (error, bool) {
	sourcePath, err := filepath.Abs(targetSpec.FileSourcePath)
	if err != nil {
		return err, false
	}
	changedPath, err = filepath.Abs(changedPath)
	if err != nil {
		return err, false
	}

	if targetSpec.FileUnpack {
		return rootAffected_pathWithin(sourcePath, changedPath)
	}

	info, err := os.Lstat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		return err, false
	}
	if info.IsDir() {
		return rootRequirementFileDirectoryOwnsPath(ctx, sourcePath, changedPath)
	}
	return rootAffected_pathWithin(sourcePath, changedPath)
}

func (roots *SafeRootsReference) FileRequirementOwners(ctx *task.ExecutionContext, changedPath string) (error, []RootFileRequirementOwner) {
	owners := []RootFileRequirementOwner{}
	err := roots.WalkVariants(ctx, RootsWalkVariantsRequest{
		OnMatch: func(ctx *task.ExecutionContext, variant *SafeRootVariantReference) (error, any) {
			if variant.Requirements == nil {
				return nil, nil
			}
			err := variant.Requirements.Walk(ctx, RootRequirementsWalkRequest{
				OnMatch: func(ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any) {
					err, requirementName, condition := rootRequirementParseName(filepath.Base(requirement.BasePath))
					if err != nil {
						return err, nil
					}
					_ = requirementName
					err, shouldInclude := rootRequirementConditionMatches(variant.Descriptor, condition)
					if err != nil || !shouldInclude {
						return err, nil
					}

					err, targetSpec := requirement.TargetSpec(ctx)
					if err != nil {
						return err, nil
					}
					if rootRequirementTargetKind(targetSpec.Kind) != RootRequirementTargetKindFile {
						return nil, nil
					}
					err, owns := rootRequirementFileOwnsPath(ctx, targetSpec, changedPath)
					if err != nil || !owns {
						return err, nil
					}
					owners = append(owners, RootFileRequirementOwner{
						RootPath: variant.Root.BasePath,
						Variant:  variant.Descriptor,
					})
					return nil, nil
				},
			})
			return err, nil
		},
	})
	return err, owners
}

package core

import (
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"strings"
)

type RootWalkAllRequirementsRequest struct {
	OnMatch func(ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any)
}

func (root *SafeRootReference) WalkAllRequirements(
	ctx *task.ExecutionContext,
	req RootWalkAllRequirementsRequest,
) error {
	dydPath := filepath.Join(root.BasePath, "dyd")

	dydEntries, err := os.ReadDir(dydPath)
	if err != nil {
		return err
	}

	for _, dydEntry := range dydEntries {
		if !dydEntry.IsDir() {
			continue
		}

		requirementsDirName := dydEntry.Name()
		if requirementsDirName != "requirements" && !strings.HasPrefix(requirementsDirName, "requirements~") {
			continue
		}

		requirementsPath := filepath.Join(dydPath, requirementsDirName)
		requirementsRef := &SafeRootRequirementsReference{
			BasePath: requirementsPath,
			Root:     root,
		}

		requirementEntries, err := os.ReadDir(requirementsPath)
		if err != nil {
			return err
		}

		for _, requirementEntry := range requirementEntries {
			if requirementEntry.IsDir() {
				continue
			}

			unsafeRequirementRef := UnsafeRootRequirementReference{
				BasePath:     filepath.Join(requirementsPath, requirementEntry.Name()),
				Requirements: requirementsRef,
			}

			err, safeRequirementRef := unsafeRequirementRef.Resolve(ctx)
			if err != nil {
				return err
			} else if safeRequirementRef == nil {
				continue
			}

			err, _ = req.OnMatch(ctx, safeRequirementRef)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

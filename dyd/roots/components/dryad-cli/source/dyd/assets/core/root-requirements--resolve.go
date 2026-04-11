package core

import (
	"dryad/task"

	"dryad/internal/os"
)

func (rootRequirements *UnsafeRootRequirementsReference) Resolve(ctx *task.ExecutionContext) (error, *SafeRootRequirementsReference) {
	var rootRequirementsExists bool
	var err error
	var safeRef SafeRootRequirementsReference

	rootRequirementsExists, err = fileExists(rootRequirements.BasePath)
	if err != nil {
		return err, nil
	}

	if !rootRequirementsExists {
		return nil, nil
	}

	safeRef = SafeRootRequirementsReference{
		BasePath: rootRequirements.BasePath,
		Root:     rootRequirements.Root,
	}

	return nil, &safeRef
}

func (rootRequirements *UnsafeRootRequirementsReference) Ensure(ctx *task.ExecutionContext) (error, *SafeRootRequirementsReference) {
	err := os.MkdirAll(rootRequirements.BasePath, os.ModePerm)
	if err != nil {
		return err, nil
	}

	return rootRequirements.Resolve(ctx)
}

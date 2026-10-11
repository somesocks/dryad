package core

import (
	"context"
	"dryad/internal/os"
)

func (rootRequirements *UnsafeRootRequirementsReference) Resolve(ctx context.Context) (error, *SafeRootRequirementsReference) {
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

func (rootRequirements *UnsafeRootRequirementsReference) Ensure(ctx context.Context) (error, *SafeRootRequirementsReference) {
	err := os.MkdirAll(rootRequirements.BasePath, os.ModePerm)
	if err != nil {
		return err, nil
	}

	return rootRequirements.Resolve(ctx)
}

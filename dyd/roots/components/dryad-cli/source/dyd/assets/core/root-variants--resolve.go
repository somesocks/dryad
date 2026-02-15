package core

import (
	"dryad/task"
)

func (rootVariants *UnsafeRootVariantsReference) Resolve(ctx *task.ExecutionContext) (error, *SafeRootVariantsReference) {
	var variantsExists bool
	var err error
	var safeRef SafeRootVariantsReference

	variantsExists, err = fileExists(rootVariants.BasePath)
	if err != nil {
		return err, nil
	}

	if !variantsExists {
		return nil, nil
	}

	safeRef = SafeRootVariantsReference{
		BasePath: rootVariants.BasePath,
		Root:     rootVariants.Root,
	}

	return nil, &safeRef
}

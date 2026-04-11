package core

import "dryad/task"

func rootBuild_selectAssetsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	rootRef := SafeRootReference{BasePath: rootPath}
	err, unsafeVariant := rootRef.VariantFromFilesystem(variantDescriptor)
	if err != nil {
		return err, ""
	}

	err, variant := unsafeVariant.Resolve(ctx)
	if err != nil || variant.Assets == nil {
		return err, ""
	}
	return nil, variant.Assets.BasePath
}

func rootBuild_selectCommandsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	rootRef := SafeRootReference{BasePath: rootPath}
	err, unsafeVariant := rootRef.VariantFromFilesystem(variantDescriptor)
	if err != nil {
		return err, ""
	}

	err, variant := unsafeVariant.Resolve(ctx)
	if err != nil || variant.Commands == nil {
		return err, ""
	}
	return nil, variant.Commands.BasePath
}

func rootBuild_selectSecretsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	rootRef := SafeRootReference{BasePath: rootPath}
	err, unsafeVariant := rootRef.VariantFromFilesystem(variantDescriptor)
	if err != nil {
		return err, ""
	}

	err, variant := unsafeVariant.Resolve(ctx)
	if err != nil || variant.Secrets == nil {
		return err, ""
	}
	return nil, variant.Secrets.BasePath
}

func rootBuild_selectDocsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	rootRef := SafeRootReference{BasePath: rootPath}
	err, unsafeVariant := rootRef.VariantFromFilesystem(variantDescriptor)
	if err != nil {
		return err, ""
	}

	err, variant := unsafeVariant.Resolve(ctx)
	if err != nil || variant.Docs == nil {
		return err, ""
	}
	return nil, variant.Docs.BasePath
}

func rootBuild_selectRequirementsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	rootRef := SafeRootReference{BasePath: rootPath}
	err, unsafeVariant := rootRef.VariantFromFilesystem(variantDescriptor)
	if err != nil {
		return err, ""
	}

	err, variant := unsafeVariant.Resolve(ctx)
	if err != nil || variant.Requirements == nil {
		return err, ""
	}
	return nil, variant.Requirements.BasePath
}

func rootBuild_selectTraitsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	rootRef := SafeRootReference{BasePath: rootPath}
	err, unsafeVariant := rootRef.VariantFromFilesystem(variantDescriptor)
	if err != nil {
		return err, ""
	}

	err, variant := unsafeVariant.Resolve(ctx)
	if err != nil || variant.Traits == nil {
		return err, ""
	}
	return nil, variant.Traits.BasePath
}

package core

import "dryad/task"

func rootBuild_selectAssetsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	err, selectedPaths := rootBuild_selectAssetsAndCommandsAndSecretsAndDocsAndRequirementsPaths(
		ctx,
		rootPath,
		variantDescriptor,
	)
	return err, selectedPaths.AssetsPath
}

func rootBuild_selectCommandsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	err, selectedPaths := rootBuild_selectAssetsAndCommandsAndSecretsAndDocsAndRequirementsPaths(
		ctx,
		rootPath,
		variantDescriptor,
	)
	return err, selectedPaths.CommandsPath
}

func rootBuild_selectSecretsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	err, selectedPaths := rootBuild_selectAssetsAndCommandsAndSecretsAndDocsAndRequirementsPaths(
		ctx,
		rootPath,
		variantDescriptor,
	)
	return err, selectedPaths.SecretsPath
}

func rootBuild_selectDocsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	err, selectedPaths := rootBuild_selectAssetsAndCommandsAndSecretsAndDocsAndRequirementsPaths(
		ctx,
		rootPath,
		variantDescriptor,
	)
	return err, selectedPaths.DocsPath
}

func rootBuild_selectRequirementsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	err, selectedPaths := rootBuild_selectAssetsAndCommandsAndSecretsAndDocsAndRequirementsPaths(
		ctx,
		rootPath,
		variantDescriptor,
	)
	return err, selectedPaths.RequirementsPath
}

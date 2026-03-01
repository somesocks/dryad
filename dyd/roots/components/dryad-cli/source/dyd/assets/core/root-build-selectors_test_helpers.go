package core

import "dryad/task"

func rootBuild_selectAssetsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	err, assetsPath, _, _, _ := rootBuild_selectAssetsAndCommandsAndSecretsAndDocsPaths(
		ctx,
		rootPath,
		variantDescriptor,
	)
	return err, assetsPath
}

func rootBuild_selectCommandsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	err, _, commandsPath, _, _ := rootBuild_selectAssetsAndCommandsAndSecretsAndDocsPaths(
		ctx,
		rootPath,
		variantDescriptor,
	)
	return err, commandsPath
}

func rootBuild_selectSecretsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	err, _, _, secretsPath, _ := rootBuild_selectAssetsAndCommandsAndSecretsAndDocsPaths(
		ctx,
		rootPath,
		variantDescriptor,
	)
	return err, secretsPath
}

func rootBuild_selectDocsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	err, _, _, _, docsPath := rootBuild_selectAssetsAndCommandsAndSecretsAndDocsPaths(
		ctx,
		rootPath,
		variantDescriptor,
	)
	return err, docsPath
}

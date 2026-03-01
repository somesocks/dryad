package core

import "dryad/task"

func rootBuild_selectAssetsPathForTest(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	err, assetsPath, _, _ := rootBuild_selectAssetsAndCommandsAndSecretsPaths(
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
	err, _, commandsPath, _ := rootBuild_selectAssetsAndCommandsAndSecretsPaths(
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
	err, _, _, secretsPath := rootBuild_selectAssetsAndCommandsAndSecretsPaths(
		ctx,
		rootPath,
		variantDescriptor,
	)
	return err, secretsPath
}

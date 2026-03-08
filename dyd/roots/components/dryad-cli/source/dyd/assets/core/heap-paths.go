package core

import (
	"path/filepath"

	"dryad/task"
)

func heapVersionDir(basePath string, version string) string {
	return filepath.Join(basePath, version)
}

func heapFilesVersionDir(basePath string) string {
	return heapVersionDir(basePath, fingerprintVersionV2)
}

func heapSecretsVersionDir(basePath string) string {
	return heapVersionDir(basePath, fingerprintVersionV2)
}

func heapStemsVersionDir(basePath string) string {
	return heapVersionDir(basePath, fingerprintVersionV2)
}

func heapSproutsVersionDir(basePath string) string {
	return heapVersionDir(basePath, fingerprintVersionV2)
}

func heapDerivationsRootsVersionDir(basePath string) string {
	return filepath.Join(basePath, "roots", fingerprintVersionV2)
}

func heapFilesFingerprintPath(ctx *task.ExecutionContext, basePath string, fingerprint string) (error, string) {
	err, depth := shedHeapFilesDepth(ctx, basePath)
	if err != nil {
		return err, ""
	}
	return heapFingerprintPath(basePath, fingerprint, depth)
}

func heapSecretsFingerprintPath(ctx *task.ExecutionContext, basePath string, fingerprint string) (error, string) {
	err, depth := shedHeapSecretsDepth(ctx, basePath)
	if err != nil {
		return err, ""
	}
	return heapFingerprintPath(basePath, fingerprint, depth)
}

func heapStemsFingerprintPath(ctx *task.ExecutionContext, basePath string, fingerprint string) (error, string) {
	err, depth := shedHeapStemsDepth(ctx, basePath)
	if err != nil {
		return err, ""
	}
	return heapFingerprintPath(basePath, fingerprint, depth)
}

func heapSproutsFingerprintPath(ctx *task.ExecutionContext, basePath string, fingerprint string) (error, string) {
	err, depth := shedHeapSproutsDepth(ctx, basePath)
	if err != nil {
		return err, ""
	}
	return heapFingerprintPath(basePath, fingerprint, depth)
}

func heapDerivationsRootsFingerprintPath(ctx *task.ExecutionContext, basePath string, fingerprint string) (error, string) {
	err, depth := shedHeapDerivationsRootsDepth(ctx, basePath)
	if err != nil {
		return err, ""
	}
	return heapFingerprintPath(filepath.Join(basePath, "roots"), fingerprint, depth)
}

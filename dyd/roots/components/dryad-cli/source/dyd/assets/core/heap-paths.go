package core

import (
	"context"
	"dryad/internal/filepath"
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

func heapFilesFingerprintPath(ctx context.Context, garden *SafeGardenReference, basePath string, fingerprint string) (error, string) {
	err, depth := shedHeapFilesDepth(ctx, safeShedReference(garden))
	if err != nil {
		return err, ""
	}
	return heapFingerprintPath(basePath, fingerprint, depth)
}

func heapSecretsFingerprintPath(ctx context.Context, garden *SafeGardenReference, basePath string, fingerprint string) (error, string) {
	err, depth := shedHeapSecretsDepth(ctx, safeShedReference(garden))
	if err != nil {
		return err, ""
	}
	return heapFingerprintPath(basePath, fingerprint, depth)
}

func heapStemsFingerprintPath(ctx context.Context, garden *SafeGardenReference, basePath string, fingerprint string) (error, string) {
	err, depth := shedHeapStemsDepth(ctx, safeShedReference(garden))
	if err != nil {
		return err, ""
	}
	return heapFingerprintPath(basePath, fingerprint, depth)
}

func heapSproutsFingerprintPath(ctx context.Context, garden *SafeGardenReference, basePath string, fingerprint string) (error, string) {
	err, depth := shedHeapSproutsDepth(ctx, safeShedReference(garden))
	if err != nil {
		return err, ""
	}
	return heapFingerprintPath(basePath, fingerprint, depth)
}

func heapDerivationsRootsFingerprintPath(ctx context.Context, garden *SafeGardenReference, basePath string, fingerprint string) (error, string) {
	err, depth := shedHeapDerivationsRootsDepth(ctx, safeShedReference(garden))
	if err != nil {
		return err, ""
	}
	return heapFingerprintPath(filepath.Join(basePath, "roots"), fingerprint, depth)
}

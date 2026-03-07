package core

import "path/filepath"

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

func heapFilesFingerprintPath(basePath string, fingerprint string) (string, error) {
	err, version, encoded := fingerprintParse(fingerprint)
	if err != nil {
		return "", err
	}
	return filepath.Join(heapVersionDir(basePath, version), encoded), nil
}

func heapSecretsFingerprintPath(basePath string, fingerprint string) (string, error) {
	err, version, encoded := fingerprintParse(fingerprint)
	if err != nil {
		return "", err
	}
	return filepath.Join(heapVersionDir(basePath, version), encoded), nil
}

func heapStemsFingerprintPath(basePath string, fingerprint string) (string, error) {
	err, version, encoded := fingerprintParse(fingerprint)
	if err != nil {
		return "", err
	}
	return filepath.Join(heapVersionDir(basePath, version), encoded), nil
}

func heapSproutsFingerprintPath(basePath string, fingerprint string) (string, error) {
	err, version, encoded := fingerprintParse(fingerprint)
	if err != nil {
		return "", err
	}
	return filepath.Join(heapVersionDir(basePath, version), encoded), nil
}

func heapDerivationsRootsFingerprintPath(basePath string, fingerprint string) (string, error) {
	err, version, encoded := fingerprintParse(fingerprint)
	if err != nil {
		return "", err
	}
	return filepath.Join(basePath, "roots", version, encoded), nil
}

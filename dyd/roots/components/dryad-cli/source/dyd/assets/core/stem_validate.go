package core

import (
	"errors"
	"os"
	"path/filepath"

	"dryad/task"
)

func StemValidate(stemPath string) (string, error) {
	var err error

	// convert relative path to absolute
	if !filepath.IsAbs(stemPath) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		stemPath = filepath.Join(wd, stemPath)
	}

	err, stemFingerprint := StemFingerprint(
		task.SERIAL_CONTEXT,
		StemFingerprintRequest{
			BasePath: stemPath,
		},
	)
	if err != nil {
		return "", err
	}

	fileFingerprintBytes, err := os.ReadFile(filepath.Join(stemPath, "dyd", "fingerprint"))
	if err != nil {
		return "", err
	}
	fileFingerprint := string(fileFingerprintBytes)

	if fileFingerprint != stemFingerprint {
		return "", errors.New("calculated fingerprint " + stemFingerprint + " doesn't match stored fingerprint " + fileFingerprint)
	}

	return stemFingerprint, nil
}

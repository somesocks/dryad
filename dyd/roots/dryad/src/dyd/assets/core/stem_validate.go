package core

import (
	"errors"
	"os"
	"path/filepath"
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

	stemFingerprint, err := StemFingerprint(
		StemFingerprintArgs{
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

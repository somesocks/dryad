package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func HeapAddSecrets(heapPath string, secretsPath string) (string, error) {
	fmt.Println("[trace] HeapAddSecrets ", heapPath, secretsPath)

	// normalize the heap path
	heapPath, err := HeapPath(heapPath)
	if err != nil {
		return "", err
	}

	// normalize the secrets path
	secretsPath, err = SecretsPath(secretsPath)
	if err != nil {
		return "", err
	}

	secretsFingerprint, err := SecretsFingerprint(SecretsFingerprintArgs{
		BasePath: secretsPath,
	})
	if err != nil {
		return "", err
	}

	// if there are no secrets, don't add to the heap
	if secretsFingerprint == "" {
		return "", nil
	}

	// check if the secrets are already in the heap
	secretsHeapPath := filepath.Join(heapPath, "secrets", secretsFingerprint)
	secretsDirExists, err := fileExists(secretsHeapPath)
	if err != nil {
		return "", err
	}

	if secretsDirExists {
		return secretsFingerprint, nil
	}

	// create the secrets dir
	err = os.MkdirAll(secretsHeapPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	var onMatch = func(path string, info fs.FileInfo) error {
		relPath, err := filepath.Rel(secretsPath, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(secretsHeapPath, relPath)

		if info.IsDir() {
			err = os.MkdirAll(destPath, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			var destFile *os.File
			destFile, err = os.Create(destPath)
			if err != nil {
				return err
			}
			defer destFile.Close()

			_, err = destFile.ReadFrom(srcFile)
			if err != nil {
				return err
			}

			err = destFile.Chmod(os.ModePerm)
			if err != nil {
				return err
			}

			err = destFile.Sync()
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = SecretsWalk(SecretsWalkArgs{
		BasePath: secretsPath,
		OnMatch:  onMatch,
	})
	if err != nil {
		return secretsFingerprint, err
	}

	return secretsFingerprint, nil
}

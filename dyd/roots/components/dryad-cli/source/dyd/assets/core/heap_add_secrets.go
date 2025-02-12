package core

import (
	fs2 "dryad/filesystem"
	"io/fs"
	"os"
	"path/filepath"
)

func HeapAddSecrets(garden *SafeGardenReference, secretsPath string) (string, error) {
	// fmt.Println("[trace] HeapAddSecrets ", heapPath, secretsPath)

	// normalize the heap path
	heapPath, err := HeapPath(garden)
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

			// heap files should be set to R-X--X--X
			err = destFile.Chmod(0o511)
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

	// now that all files are added, sweep through in a second pass and make directories read-only
	err = fs2.Walk(fs2.WalkRequest{
		BasePath: secretsHeapPath,
		MatchInclude: func(path string, info fs.FileInfo) (bool, error) {
			return info.IsDir(), nil
		},
		OnMatch: func(path string, info fs.FileInfo) error {
			dir, err := os.Open(path)
			if err != nil {
				return err
			}
			defer dir.Close()

			// heap files should be set to R-X--X--X
			err = dir.Chmod(0o511)
			if err != nil {
				return err
			}

			return nil
		},
	})

	if err != nil {
		return secretsFingerprint, err
	}

	return secretsFingerprint, nil
}

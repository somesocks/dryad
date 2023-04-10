package core

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

func _readFile(filePath string) (string, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// HeapAddStem takes a stem in a directory, and adds it to the heap.
// the heap path is normalized before adding
func HeapAddStem(heapPath string, stemPath string) (string, error) {
	fmt.Println("[trace] HeapAddStem", heapPath, stemPath)

	// normalize the heap path
	heapPath, err := HeapPath(heapPath)
	if err != nil {
		return "", err
	}

	gardenFilesPath := filepath.Join(heapPath, "files")
	gardenStemsPath := filepath.Join(heapPath, "stems")

	stemFingerprint, err := _readFile(filepath.Join(stemPath, "dyd", "fingerprint"))
	if err != nil {
		return "", err
	}

	finalStemPath := filepath.Join(gardenStemsPath, stemFingerprint)

	// check to see if the stem already exists in the garden
	stemExists, err := fileExists(finalStemPath)
	if err != nil {
		return "", err
	}

	if !stemExists {
		err = os.MkdirAll(finalStemPath, fs.ModePerm)
		if err != nil {
			return "", err
		}

		// walk the packed root files and copy them into the garden heap
		err = StemWalk(
			StemWalkArgs{
				BasePath: stemPath,
				OnMatch: func(srcPath string, info fs.FileInfo) error {
					// fmt.Println("HeapAddStem stemwalk", srcPath)

					var err error

					if info.IsDir() {
						return nil
					}

					relPath, err := filepath.Rel(stemPath, srcPath)
					if err != nil {
						return err
					}

					destPath := filepath.Join(finalStemPath, relPath)
					err = os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
					if err != nil {
						return err
					}

					fileFingerprint, err := HeapAddFile(gardenFilesPath, srcPath)
					if err != nil {
						return err
					}

					fileHeapPath := filepath.Join(gardenFilesPath, fileFingerprint)

					relativeFilePath, err := filepath.Rel(filepath.Dir(destPath), fileHeapPath)
					if err != nil {
						return err
					}

					err = os.Symlink(relativeFilePath, destPath)
					if err != nil {
						return err
					}

					return nil
				},
			},
		)
		if err != nil {
			return "", err
		}

		// walk the dependencies and convert them to symlinks
		dependenciesPath := filepath.Join(finalStemPath, "dyd", "stems")
		dependencies, err := filepath.Glob(filepath.Join(dependenciesPath, "*"))
		if err != nil {
			return "", err
		}

		for _, dependencyPath := range dependencies {
			targetFingerprintFile := filepath.Join(dependencyPath, "dyd", "fingerprint")
			targetFingerprintBytes, err := ioutil.ReadFile(targetFingerprintFile)
			if err != nil {
				return "", err
			}
			targetFingerprint := string(targetFingerprintBytes)

			dependencyGardenPath := filepath.Join(gardenStemsPath, targetFingerprint)
			relPath, err := filepath.Rel(dependenciesPath, dependencyGardenPath)
			if err != nil {
				return "", err
			}

			err = os.RemoveAll(dependencyPath)
			if err != nil {
				return "", err
			}

			err = os.Symlink(relPath, dependencyPath)
			if err != nil {
				return "", err
			}
		}

		secretsFingerprintPath := filepath.Join(finalStemPath, "dyd", "secrets-fingerprint")

		hasSecrets, err := fileExists(secretsFingerprintPath)
		if err != nil {
			return "", err
		}

		if hasSecrets {
			secretsFingerprint, err := HeapAddSecrets(heapPath, stemPath)
			if err != nil {
				return "", err
			}

			secretsMountPoint := filepath.Join(finalStemPath, "dyd", "secrets")
			secretsHeapPath := filepath.Join(heapPath, "secrets", secretsFingerprint)

			relativeLink, err := filepath.Rel(
				filepath.Dir(secretsMountPoint),
				secretsHeapPath,
			)
			if err != nil {
				return "", err
			}

			err = os.Symlink(relativeLink, secretsMountPoint)
			if err != nil {
				return "", err
			}

		}

	}
	return finalStemPath, nil
}

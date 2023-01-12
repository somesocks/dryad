package core

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

func RootBuild(rootPath string) (string, error) {

	// fmt.Println("RootBuild ", rootPath)

	// sanitize the root path
	rootPath, err := RootPath(rootPath)
	if err != nil {
		return "", err
	}

	// prepare a workspace
	workspacePath, err := os.MkdirTemp("", "dryad-build-*")
	if err != nil {
		return "", err
	}

	// shallow-clone into the workspace so we have something to modify
	err = StemWalk(rootPath, func(srcPath string, info fs.FileInfo, err error) error {
		// fmt.Println("RootBuild shallow clone StemWalk ", rootPath, " ", srcPath)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(rootPath, srcPath)
		if err != nil {
			return err
		}

		destPath := filepath.Join(workspacePath, relPath)
		err = os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
		if err != nil {
			return err
		}

		err = os.Symlink(srcPath, destPath)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	// walk through the dependencies, build them, and add the fingerprint as a dependency
	rootsPath := filepath.Join(rootPath, "dyd", "roots")

	dependencies, err := filepath.Glob(filepath.Join(rootsPath, "*"))
	if err != nil {
		return "", err
	}

	for _, dependencyPath := range dependencies {
		dependencyFingerprint, err := RootBuild(dependencyPath)
		if err != nil {
			return "", err
		}
		dependencyName := filepath.Base(dependencyPath)
		targetDepPath := filepath.Join(workspacePath, "dyd", "stems", dependencyName)
		targetDydDir := filepath.Join(targetDepPath, "dyd")
		targetFingerprintFile := filepath.Join(targetDydDir, "fingerprint")
		err = os.MkdirAll(targetDydDir, fs.ModePerm)
		if err != nil {
			return "", err
		}
		err = os.WriteFile(targetFingerprintFile, []byte(dependencyFingerprint), fs.ModePerm)
		if err != nil {
			return "", err
		}

	}

	// fmt.Println("RootBuild StemFingerprint", rootPath, " ", workspacePath)

	rootFingerprint, err := StemFingerprint(workspacePath)
	if err != nil {
		return "", err
	}

	// write out the fingerprint file
	err = os.WriteFile(filepath.Join(workspacePath, "dyd", "fingerprint"), []byte(rootFingerprint), fs.ModePerm)
	if err != nil {
		return "", err
	}

	// check to see if the stem already exists in the garden
	gardenPath, err := GardenPath(rootPath)
	if err != nil {
		return "", err
	}

	gardenHeapPath := filepath.Join(gardenPath, "dyd", "heap")
	gardenStemsPath := filepath.Join(gardenPath, "dyd", "stems")

	finalStemPath := filepath.Join(gardenStemsPath, rootFingerprint)

	stemExists, err := fileExists(finalStemPath)
	if err != nil {
		return "", err
	}

	if !stemExists {
		err = os.MkdirAll(finalStemPath, fs.ModePerm)
		if err != nil {
			return "", err
		}

		// walk the packed root files and copy them into the garden stems
		err = StemWalk(workspacePath, func(srcPath string, info fs.FileInfo, err error) error {
			// fmt.Println("StemWalk pack into garden ", rootPath, " ", srcPath)
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(workspacePath, srcPath)
			if err != nil {
				return err
			}

			destPath := filepath.Join(finalStemPath, relPath)
			err = os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
			if err != nil {
				return err
			}

			fileFingerprint, err := HeapAdd(gardenHeapPath, srcPath)
			if err != nil {
				return err
			}

			fileHeapPath := filepath.Join(gardenHeapPath, fileFingerprint)

			relativeFilePath, err := filepath.Rel(filepath.Dir(destPath), fileHeapPath)
			if err != nil {
				return err
			}

			err = os.Symlink(relativeFilePath, destPath)
			if err != nil {
				return err
			}

			return nil
		})
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

	}
	// fmt.Println("RootBuild ", rootPath, " ", rootFingerprint)

	return rootFingerprint, nil
}

package core

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

	"path/filepath"
)

// stage 0 - shallow-clone the root into a working directory,
// so we have something to modify, and a place to generate the fingerprint
func rootBuild_stage0(rootPath string, workspacePath string) error {
	err := StemWalk(rootPath, func(srcPath string, info fs.FileInfo, err error) error {
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
	return err
}

// stage 1 - walk through the root dependencies,
// and add the fingerprint as a dependency
func rootBuild_stage1(context BuildContext, rootPath string, workspacePath string) error {
	// walk through the dependencies, build them, and add the fingerprint as a dependency
	rootsPath := filepath.Join(rootPath, "dyd", "roots")

	dependencies, err := filepath.Glob(filepath.Join(rootsPath, "*"))
	if err != nil {
		return err
	}

	for _, dependencyPath := range dependencies {
		dependencyFingerprint, err := RootBuild(context, dependencyPath)
		if err != nil {
			return err
		}
		dependencyName := filepath.Base(dependencyPath)
		targetDepPath := filepath.Join(workspacePath, "dyd", "stems", dependencyName)
		targetDydDir := filepath.Join(targetDepPath, "dyd")
		targetFingerprintFile := filepath.Join(targetDydDir, "fingerprint")
		err = os.MkdirAll(targetDydDir, fs.ModePerm)
		if err != nil {
			return err
		}
		err = os.WriteFile(targetFingerprintFile, []byte(dependencyFingerprint), fs.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

// stage 2 - generate the artificial links to all executable stems for the path
func rootBuild_stage2(workspacePath string) error {
	pathPath := filepath.Join(workspacePath, "dyd", "path")

	err := os.RemoveAll(pathPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(pathPath, fs.ModePerm)
	if err != nil {
		return err
	}

	// walk through the dependencies, build them, and add the fingerprint as a dependency
	dependenciesPath := filepath.Join(workspacePath, "dyd", "stems")

	dependencies, err := filepath.Glob(filepath.Join(dependenciesPath, "*"))
	if err != nil {
		return err
	}

	for _, dependencyPath := range dependencies {
		basename := filepath.Base(dependencyPath)

		baseTemplate := fmt.Sprintf(
			"#!/usr/bin/env sh\nPACKAGE=\"./dyd/stems/%s/\" PATH=\"../stems/%s/dyd/path\" ../stems/%s/dyd/main",
			basename,
			basename,
			basename,
		)

		err = os.WriteFile(
			filepath.Join(pathPath, basename),
			[]byte(baseTemplate),
			fs.ModePerm,
		)
		if err != nil {
			return err
		}

	}

	return nil
}

// stage 3 - generate the fingerprint for the newly-constructed root,
// and write it out to the fingerprint file
func rootBuild_stage3(rootPath string, workspacePath string) (string, error) {

	rootFingerprint, err := StemFingerprint(workspacePath)
	if err != nil {
		return "", err
	}

	// write out the fingerprint file
	err = os.WriteFile(filepath.Join(workspacePath, "dyd", "fingerprint"), []byte(rootFingerprint), fs.ModePerm)
	if err != nil {
		return "", err
	}

	return rootFingerprint, nil
}

// stage 4 - check the garden to see if the stem exists,
// and add it if it doesn't
func rootBuild_stage4(gardenPath string, workspacePath string, rootFingerprint string) (string, error) {

	gardenFilesPath := filepath.Join(gardenPath, "dyd", "heap", "files")
	gardenStemsPath := filepath.Join(gardenPath, "dyd", "heap", "stems")

	finalStemPath := filepath.Join(gardenStemsPath, rootFingerprint)

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
		err = StemWalk(workspacePath, func(srcPath string, info fs.FileInfo, err error) error {
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

			fileFingerprint, err := HeapAdd(gardenFilesPath, srcPath)
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
	return finalStemPath, nil
}

// stage 5 - execute the root to build its stem,
func rootBuild_stage5(rootStemPath string, stemBuildPath string, rootFingerprint string) (string, error) {
	var err error

	err = StemInit(stemBuildPath)
	if err != nil {
		return "", err
	}

	err = StemExec(rootStemPath, stemBuildPath)
	if err != nil {
		return "", err
	}

	// write out the source file
	err = os.WriteFile(filepath.Join(stemBuildPath, "dyd", "traits", "source"), []byte(rootFingerprint), fs.ModePerm)
	if err != nil {
		return "", err
	}

	stemBuildFingerprint, err := StemFingerprint(stemBuildPath)
	if err != nil {
		return "", err
	}

	// write out the fingerprint file
	err = os.WriteFile(filepath.Join(stemBuildPath, "dyd", "fingerprint"), []byte(stemBuildFingerprint), fs.ModePerm)
	if err != nil {
		return "", err
	}

	return stemBuildFingerprint, err
}

// stage 6 - pack the dervied stem into the heap and garden
func rootBuild_stage6(gardenPath string, sourcePath string, stemFingerprint string) (string, error) {

	gardenFilesPath := filepath.Join(gardenPath, "dyd", "heap", "files")
	gardenStemsPath := filepath.Join(gardenPath, "dyd", "heap", "stems")

	finalStemPath := filepath.Join(gardenStemsPath, stemFingerprint)

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
		err = StemWalk(sourcePath, func(srcPath string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(sourcePath, srcPath)
			if err != nil {
				return err
			}

			destPath := filepath.Join(finalStemPath, relPath)
			err = os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
			if err != nil {
				return err
			}

			fileFingerprint, err := HeapAdd(gardenFilesPath, srcPath)
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
	return finalStemPath, nil
}

func RootBuild(context BuildContext, rootPath string) (string, error) {

	// sanitize the root path
	rootPath, err := RootPath(rootPath)
	if err != nil {
		return "", err
	}

	absRootPath, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return "", err
	}

	// check if the root is already present in the context
	rootFingerprint, contextHasRootFingerprint := context.RootFingerprints[absRootPath]
	if contextHasRootFingerprint {
		return rootFingerprint, nil
	}

	// check to see if the stem already exists in the garden
	gardenPath, err := GardenPath(rootPath)
	if err != nil {
		return "", err
	}

	// prepare a workspace
	workspacePath, err := os.MkdirTemp("", "dryad-build-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(workspacePath)

	err = rootBuild_stage0(rootPath, workspacePath)
	if err != nil {
		return "", err
	}

	err = rootBuild_stage1(context, rootPath, workspacePath)
	if err != nil {
		return "", err
	}

	err = rootBuild_stage2(workspacePath)
	if err != nil {
		return "", err
	}

	rootFingerprint, err = rootBuild_stage3(rootPath, workspacePath)
	if err != nil {
		return "", err
	}

	finalStemPath, err := rootBuild_stage4(gardenPath, workspacePath, rootFingerprint)
	if err != nil {
		return "", err
	}

	var stemBuildFingerprint string

	// if the derivation link already exists,
	// then return it directly
	derivationsPath := filepath.Join(gardenPath, "dyd", "heap", "derivations", rootFingerprint)
	derivationFileExists, err := fileExists(derivationsPath)
	if err != nil {
		return "", err
	}

	if derivationFileExists {
		derivationsFingerprintFile := filepath.Join(derivationsPath, "dyd", "fingerprint")
		derivationsFingerprintBytes, err := ioutil.ReadFile(derivationsFingerprintFile)
		if err != nil {
			return "", err
		}
		derivationsFingerprint := string(derivationsFingerprintBytes)

		// add the built fingerprint to the context
		context.RootFingerprints[absRootPath] = derivationsFingerprint

		stemBuildFingerprint = derivationsFingerprint

	} else {
		// otherwise run the root in a build env
		stemBuildPath, err := os.MkdirTemp("", "dryad-build-*")
		if err != nil {
			return "", err
		}
		defer os.RemoveAll(stemBuildPath)

		stemBuildFingerprint, err := rootBuild_stage5(finalStemPath, stemBuildPath, rootFingerprint)
		if err != nil {
			return "", err
		}

		finalStemPath, err = rootBuild_stage6(gardenPath, stemBuildPath, stemBuildFingerprint)
		if err != nil {
			return "", err
		}

		// add the built fingerprint to the context
		context.RootFingerprints[absRootPath] = stemBuildFingerprint

		// add the derivation link
		derivationsLinkPath, err := filepath.Rel(
			filepath.Dir(derivationsPath),
			finalStemPath,
		)
		if err != nil {
			return "", err
		}
		err = os.RemoveAll(derivationsPath)
		if err != nil {
			return "", err
		}
		err = os.Symlink(derivationsLinkPath, derivationsPath)
		if err != nil {
			return "", err
		}

	}

	relRootPath, err := filepath.Rel(
		filepath.Join(gardenPath, "dyd", "roots"),
		rootPath,
	)
	if err != nil {
		return "", err
	}

	sproutPath := filepath.Join(gardenPath, "dyd", "sprouts", relRootPath)
	sproutParent := filepath.Dir(sproutPath)
	relSproutLink, err := filepath.Rel(
		sproutParent,
		filepath.Join(gardenPath, "dyd", "heap", "stems", stemBuildFingerprint),
	)
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(sproutParent, fs.ModePerm)
	if err != nil {
		return "", err
	}

	err = os.Remove(sproutPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	err = os.Symlink(relSproutLink, sproutPath)
	if err != nil {
		return "", err
	}

	return stemBuildFingerprint, nil
}

package core

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func rootBuild_pathStub(depname string) string {
	return `#!/usr/bin/env sh
set -eu
STEM_PATH="$(dirname $0)/../stems/$(basename $0)"
PATH="$STEM_PATH/dyd/path:$PATH" \
DYD_STEM="$STEM_PATH" \
"$STEM_PATH"/dyd/main $@
`
}

// stage 0 - build a shallow partial clone of the root into a working directory,
// so we can build it into a stem
func rootBuild_stage0(rootPath string, workspacePath string) error {
	// fmt.Println("rootBuild_stage0 ", rootPath, " ", workspacePath)

	rootPath, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(
		filepath.Join(workspacePath, "dyd"),
		os.ModePerm,
	)
	if err != nil {
		return err
	}

	err = os.Symlink(
		filepath.Join(rootPath, "dyd", "main"),
		filepath.Join(workspacePath, "dyd", "main"),
	)
	if err != nil {
		return err
	}

	err = os.Symlink(
		filepath.Join(rootPath, "dyd", "assets"),
		filepath.Join(workspacePath, "dyd", "assets"),
	)
	if err != nil {
		return err
	}

	err = os.Symlink(
		filepath.Join(rootPath, "dyd", "traits"),
		filepath.Join(workspacePath, "dyd", "traits"),
	)
	if err != nil {
		return err
	}

	return nil
}

// stage 1 - walk through the root dependencies,
// and add the fingerprint as a dependency
func rootBuild_stage1(context BuildContext, rootPath string, workspacePath string) error {
	// fmt.Println("rootBuild_stage1 ", rootPath, " ", workspacePath)

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
	// fmt.Println("rootBuild_stage2 ", workspacePath)

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

		baseTemplate := rootBuild_pathStub(basename)

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

var _RE_SANITIZE_ENV, _ = regexp.Compile(`[^a-zA-Z0-9_]+`)

func rootBuild_sanitizeEnvName(env string) string {
	return strings.ToUpper(
		_RE_SANITIZE_ENV.ReplaceAllString(env, "_"),
	)
}

// stage 3 - read the root secrets,
// generate the fingerprint,
// and add the secrets to the heap
func rootBuild_stage3(rootPath string, workspacePath string) (string, error) {
	// fmt.Println("rootBuild_stage3 ", rootPath, " ", workspacePath)
	secretsFingerprint, err := HeapAddSecrets(rootPath, rootPath)
	if err != nil {
		return secretsFingerprint, err
	}

	// don't create a fingerprint if there are no secrets
	if secretsFingerprint == "" {
		return secretsFingerprint, nil
	}

	// write out the secrets fingerprint
	err = os.WriteFile(
		filepath.Join(workspacePath, "dyd", "traits", "secrets-fingerprint"),
		[]byte(secretsFingerprint),
		fs.ModePerm,
	)
	if err != nil {
		return secretsFingerprint, err
	}

	return secretsFingerprint, nil
}

// stage 4 - generate the fingerprint for the newly-constructed root,
// and write it out to the fingerprint file
func rootBuild_stage4(rootPath string, workspacePath string) (string, error) {
	// fmt.Println("rootBuild_stage4 ", rootPath, " ", workspacePath)

	rootFingerprint, err := StemFingerprint(
		StemFingerprintArgs{
			BasePath: workspacePath,
		},
	)
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

// stage 5 - check the garden to see if the stem exists,
// and add it if it doesn't
func rootBuild_stage5(gardenPath string, workspacePath string, rootFingerprint string) (string, error) {
	// fmt.Println("rootBuild_stage5 ", workspacePath)

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
		err = StemWalk(
			StemWalkArgs{
				BasePath: workspacePath,
				OnMatch: func(srcPath string, info fs.FileInfo) error {
					var err error

					if info.IsDir() {
						return nil
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
	}
	return finalStemPath, nil
}

// stage 6 - execute the root to build its stem,
func rootBuild_stage6(rootStemPath string, stemBuildPath string, rootFingerprint string) (string, error) {
	// fmt.Println("rootBuild_stage6 ", rootStemPath)

	var err error

	err = StemInit(stemBuildPath)
	if err != nil {
		return "", err
	}
	err = StemExec(StemExecRequest{
		StemPath:   rootStemPath,
		Env:        nil,
		Args:       []string{stemBuildPath},
		JoinStdout: false,
	})
	if err != nil {
		return "", err
	}

	// write out the source file
	sourceFile := filepath.Join(stemBuildPath, "dyd", "traits", "root-fingerprint")
	sourceFileExists, err := fileExists(sourceFile)
	if err != nil {
		return "", err
	}
	if sourceFileExists {
		err = os.Remove(sourceFile)
		if err != nil {
			return "", err
		}
	}

	err = os.WriteFile(
		sourceFile,
		[]byte(rootFingerprint),
		fs.ModePerm,
	)
	if err != nil {
		return "", err
	}

	// write out the path files
	pathPath := filepath.Join(stemBuildPath, "dyd", "path")

	err = os.RemoveAll(pathPath)
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(pathPath, fs.ModePerm)
	if err != nil {
		return "", err
	}

	// walk through the dependencies, build them, and add the fingerprint as a dependency
	dependenciesPath := filepath.Join(stemBuildPath, "dyd", "stems", "*")

	dependencies, err := filepath.Glob(dependenciesPath)
	if err != nil {
		return "", err
	}

	for _, dependencyPath := range dependencies {
		basename := filepath.Base(dependencyPath)

		baseTemplate := rootBuild_pathStub(basename)

		err = os.WriteFile(
			filepath.Join(pathPath, basename),
			[]byte(baseTemplate),
			fs.ModePerm,
		)
		if err != nil {
			return "", err
		}

	}

	stemBuildFingerprint, err := StemFingerprint(
		StemFingerprintArgs{
			BasePath: stemBuildPath,
		},
	)
	if err != nil {
		return "", err
	}

	// write out the fingerprint file
	fingerprintFile := filepath.Join(stemBuildPath, "dyd", "fingerprint")
	fingerprintFileExists, err := fileExists(fingerprintFile)
	if err != nil {
		return "", err
	}
	if fingerprintFileExists {
		err = os.Remove(fingerprintFile)
		if err != nil {
			return "", err
		}
	}
	err = os.WriteFile(fingerprintFile, []byte(stemBuildFingerprint), fs.ModePerm)
	if err != nil {
		return "", err
	}

	return stemBuildFingerprint, err
}

// stage 7 - pack the dervied stem into the heap and garden
func rootBuild_stage7(gardenPath string, sourcePath string, stemFingerprint string) (string, error) {
	// fmt.Println("rootBuild_stage7", gardenPath, " ", sourcePath, " ", stemFingerprint)

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
		err = StemWalk(
			StemWalkArgs{
				BasePath: sourcePath,
				OnMatch: func(srcPath string, info fs.FileInfo) error {
					var err error

					if info.IsDir() {
						return nil
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
	}

	// sanity check to verify that the built stem is correct
	_, err = StemValidate(finalStemPath)
	if err != nil {
		return "", err
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

	// check to see if the stem already exists in the garden
	gardenPath, err := GardenPath(rootPath)
	if err != nil {
		return "", err
	}

	relRootPath, err := filepath.Rel(
		filepath.Join(gardenPath, "dyd", "roots"),
		absRootPath,
	)
	if err != nil {
		return "", err
	}

	// check if the root is already present in the context
	rootFingerprint, contextHasRootFingerprint := context.RootFingerprints[absRootPath]
	if contextHasRootFingerprint {
		return rootFingerprint, nil
	}

	fmt.Println("[info] dryad checking root " + relRootPath)

	// prepare a workspace
	workspacePath, err := os.MkdirTemp("", "dryad-*")
	if err != nil {
		return "", err
	}
	// defer os.RemoveAll(workspacePath)

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

	_, err = rootBuild_stage3(rootPath, workspacePath)
	if err != nil {
		return "", err
	}

	rootFingerprint, err = rootBuild_stage4(rootPath, workspacePath)
	if err != nil {
		return "", err
	}

	finalStemPath, err := rootBuild_stage5(gardenPath, workspacePath, rootFingerprint)
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

		stemBuildFingerprint = derivationsFingerprint

		// add the built fingerprint to the context
		context.RootFingerprints[absRootPath] = derivationsFingerprint

	} else {
		fmt.Println("[info] dryad building root " + relRootPath)

		// otherwise run the root in a build env
		stemBuildPath, err := os.MkdirTemp("", "dryad-*")
		if err != nil {
			return "", err
		}
		// defer os.RemoveAll(stemBuildPath)

		stemBuildFingerprint, err = rootBuild_stage6(finalStemPath, stemBuildPath, rootFingerprint)
		if err != nil {
			return "", err
		}

		finalStemPath, err = rootBuild_stage7(gardenPath, stemBuildPath, stemBuildFingerprint)
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

		fmt.Println("[info] dryad done building root " + relRootPath)
	}

	sproutPath := filepath.Join(gardenPath, "dyd", "sprouts", relRootPath)
	sproutParent := filepath.Dir(sproutPath)
	sproutHeapPath := filepath.Join(gardenPath, "dyd", "heap", "stems", stemBuildFingerprint)
	relSproutLink, err := filepath.Rel(
		sproutParent,
		sproutHeapPath,
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
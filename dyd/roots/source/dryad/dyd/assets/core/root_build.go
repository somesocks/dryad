package core

import (
	fs2 "dryad/filesystem"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/rs/zerolog/log"
)

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

	readmePath := filepath.Join(rootPath, "dyd", "readme")
	exists, err := fileExists(readmePath)
	if err != nil {
		return err
	}
	if exists {
		err = os.Symlink(
			readmePath,
			filepath.Join(workspacePath, "dyd", "readme"),
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "assets"))
	if err != nil {
		return err
	}
	if exists {
		err = os.Symlink(
			filepath.Join(rootPath, "dyd", "assets"),
			filepath.Join(workspacePath, "dyd", "assets"),
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "commands"))
	if err != nil {
		return err
	}
	if exists {
		err = os.Symlink(
			filepath.Join(rootPath, "dyd", "commands"),
			filepath.Join(workspacePath, "dyd", "commands"),
		)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(filepath.Join(workspacePath, "dyd", "dependencies"), fs.ModePerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(workspacePath, "dyd", "requirements"), fs.ModePerm)
	if err != nil {
		return err
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "secrets"))
	if err != nil {
		return err
	}
	if exists {
		err = os.Symlink(
			filepath.Join(rootPath, "dyd", "secrets"),
			filepath.Join(workspacePath, "dyd", "secrets"),
		)
		if err != nil {
			return err
		}
	}

	exists, err = fileExists(filepath.Join(rootPath, "dyd", "traits"))
	if err != nil {
		return err
	}
	if exists {
		err = os.Symlink(
			filepath.Join(rootPath, "dyd", "traits"),
			filepath.Join(workspacePath, "dyd", "traits"),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// stage 1 - walk through the root dependencies,
// and add the fingerprint as a dependency
func rootBuild_stage1(
	context BuildContext,
	rootPath string,
	workspacePath string,
	gardenPath string,
) error {

	// walk through the dependencies, build them, and add the fingerprint as a dependency
	rootsPath := filepath.Join(rootPath, "dyd", "requirements")

	dependencies, err := filepath.Glob(filepath.Join(rootsPath, "*"))
	if err != nil {
		return err
	}

	for _, dependencyPath := range dependencies {

		dependencyFingerprint, err := RootBuild(context, dependencyPath)
		if err != nil {
			return err
		}

		dependencyHeapPath := filepath.Join(gardenPath, "dyd", "heap", "stems", dependencyFingerprint)

		dependencyName := filepath.Base(dependencyPath)

		targetDepPath := filepath.Join(workspacePath, "dyd", "dependencies", dependencyName)

		err = os.Symlink(dependencyHeapPath, targetDepPath)

		if err != nil {
			return err
		}
	}

	return nil
}

// stage 2 - generate the artificial links to all executable stems for the path
func rootBuild_stage2(workspacePath string) error {
	err := rootBuild_pathPrepare(workspacePath)
	if err != nil {
		return err
	}
	err = rootBuild_requirementsPrepare(workspacePath)
	if err != nil {
		return err
	}
	return nil
}

// stage 3 - finalize the stem by generating fingerprints,
func rootBuild_stage3(rootPath string, workspacePath string) (string, error) {
	stemFingerprint, err := stemFinalize(workspacePath)
	return stemFingerprint, err
}

// stage 5 - check the garden to see if the stem exists,
// and add it if it doesn't
func rootBuild_stage5(gardenPath string, workspacePath string, rootFingerprint string) (string, error) {
	// fmt.Println("[trace] rootBuild_stage5", gardenPath, workspacePath, rootFingerprint)
	return HeapAddStem(gardenPath, workspacePath)
}

// stage 6 - execute the root to build its stem,
func rootBuild_stage6(rootStemPath string, stemBuildPath string, rootFingerprint string) (string, error) {
	// fmt.Println("rootBuild_stage6 ", rootStemPath)

	var err error

	err = StemInit(stemBuildPath)
	if err != nil {
		return "", err
	}
	err = StemRun(StemRunRequest{
		StemPath: rootStemPath,
		Env: map[string]string{
			"DYD_BUILD": stemBuildPath,
		},
		Args:       []string{stemBuildPath},
		JoinStdout: false,
	})
	if err != nil {
		return "", err
	}

	// prepare the path
	err = rootBuild_pathPrepare(stemBuildPath)
	if err != nil {
		return "", err
	}

	// prepare the requirements dir
	err = rootBuild_requirementsPrepare(stemBuildPath)
	if err != nil {
		return "", err
	}

	stemBuildFingerprint, err := stemFinalize(stemBuildPath)
	if err != nil {
		return "", err
	}

	return stemBuildFingerprint, err
}

// stage 7 - pack the dervied stem into the heap and garden
func rootBuild_stage7(gardenPath string, sourcePath string, stemFingerprint string) (string, error) {
	// fmt.Println("[trace] rootBuild_stage7", gardenPath, sourcePath, stemFingerprint)
	return HeapAddStem(gardenPath, sourcePath)
}

func RootBuild(context BuildContext, rootPath string) (string, error) {
	// fmt.Println("[trace] RootBuild", context, rootPath)

	// sanitize the root path
	rootPath, err := RootPath(rootPath)
	if err != nil {
		return "", err
	}
	// fmt.Println("[trace] RootBuild rootPath", rootPath)

	absRootPath, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return "", err
	}
	// fmt.Println("[trace] RootBuild absRootPath", absRootPath)

	// check to see if the stem already exists in the garden
	gardenPath, err := GardenPath(rootPath)
	if err != nil {
		return "", err
	}
	// fmt.Println("[trace] RootBuild gardenPath", gardenPath)

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

	log.Info().Msg("dryad checking root " + relRootPath)

	// prepare a workspace
	workspacePath, err := os.MkdirTemp("", "dryad-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(workspacePath)

	err = rootBuild_stage0(rootPath, workspacePath)
	if err != nil {
		return "", err
	}

	err = rootBuild_stage1(context, rootPath, workspacePath, gardenPath)
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

	finalStemPath, err := rootBuild_stage5(gardenPath, workspacePath, rootFingerprint)
	if err != nil {
		return "", err
	}

	isUnstableRoot, err := fileExists(filepath.Join(finalStemPath, "dyd", "traits", "unstable"))
	if err != nil {
		return "", err
	}

	var stemBuildFingerprint string

	var derivationsPath string = ""
	var derivationFileExists bool = false

	if !isUnstableRoot {
		// if the derivation link already exists,
		// then return it directly
		derivationsPath = filepath.Join(gardenPath, "dyd", "heap", "derivations", rootFingerprint)
		derivationFileExists, err = fileExists(derivationsPath)
		if err != nil {
			return "", err
		}
	}

	if derivationFileExists {
		// fmt.Println("[trace] derivationFileExists " + derivationsPath)
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
		log.Info().Msg("dryad building root " + relRootPath)

		// otherwise run the root in a build env
		stemBuildPath, err := os.MkdirTemp("", "dryad-*")
		if err != nil {
			return "", err
		}
		defer os.RemoveAll(stemBuildPath)

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

		if !isUnstableRoot {
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

		log.Info().Msg("dryad done building root " + relRootPath)
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

	// fmt.Println("[debug] building sprout parent")
	err = fs2.MkDir(sproutParent, fs.ModePerm)
	if err != nil {
		return "", err
	}

	// fmt.Println("[debug] setting write permission on sprout parent")
	err = os.Chmod(sproutParent, 0o711)
	if err != nil {
		return "", err
	}

	tmpSproutPath := sproutPath + ".tmp"
	// fmt.Println("[debug] adding temporary sprout link")
	err = os.Symlink(relSproutLink, tmpSproutPath)
	if err != nil {
		return "", err
	}

	// fmt.Println("[debug] renaming sprout link", sproutPath)
	err = os.Rename(tmpSproutPath, sproutPath)
	if err != nil {
		return "", err
	}

	// fmt.Println("[debug] setting read permissions on sprout parent")
	err = os.Chmod(sproutParent, 0o511)
	if err != nil {
		return "", err
	}

	return stemBuildFingerprint, nil
}

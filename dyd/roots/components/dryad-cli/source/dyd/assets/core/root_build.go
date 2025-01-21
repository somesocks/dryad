package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"

	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type RootBuildRequest struct {
	Context BuildContext
	RootPath string
}

func RootBuild(ctx *task.ExecutionContext, req RootBuildRequest) (error, string) {
	var rootPath string = req.RootPath
	var context BuildContext = req.Context

	// fmt.Println("[trace] RootBuild", context, rootPath)

	// sanitize the root path
	rootPath, err := RootPath(rootPath, "")
	zlog.Debug().
		Str("rootPath", rootPath).
		Msg("RootBuild/rootPath")
	if err != nil {
		return err, ""
	}

	// check to see if the stem already exists in the garden
	gardenPath, err := GardenPath(rootPath)
	zlog.Debug().
		Str("gardenPath", gardenPath).
		Msg("RootBuild/gardenPath")
	if err != nil {
		return err, ""
	}
	// fmt.Println("[trace] RootBuild gardenPath", gardenPath)

	relRootPath, err := filepath.Rel(
		filepath.Join(gardenPath, "dyd", "roots"),
		rootPath,
	)
	zlog.Debug().
		Str("relRootPath", relRootPath).
		Msg("RootBuild/relRootPath")
	if err != nil {
		return err, ""
	}

	// check if the root is already present in the context
	context.FingerprintsMutex.Lock()
	rootFingerprint, contextHasRootFingerprint := context.Fingerprints[rootPath]
	context.FingerprintsMutex.Unlock()
	if contextHasRootFingerprint {
		return nil, rootFingerprint
	}

	zlog.Info().
		Str("path", relRootPath).
		Msg("root build - verifying root")

	// prepare a workspace
	workspacePath, err := os.MkdirTemp("", "dryad-*")
	if err != nil {
		return err, ""
	}
	defer dydfs.RemoveAll(ctx, workspacePath)

	err, _ = rootBuild_stage0(
		ctx,
		rootBuild_stage0_request{
			RootPath: rootPath,
			WorkspacePath: workspacePath,
		},
	)
	if err != nil {
		return err, ""
	}

	err, _ = rootBuild_stage1(
		ctx,
		rootBuild_stage1_request{
			Context: context,
			RootPath: rootPath,
			WorkspacePath: workspacePath,
			GardenPath: gardenPath,
		},
	)
	if err != nil {
		return err, ""
	}

	err, _ = rootBuild_stage2(
		ctx,
		rootBuild_stage2_request{
			Context: context,
			RootPath: rootPath,
			WorkspacePath: workspacePath,
			GardenPath: gardenPath,
		},
	)
	if err != nil {
		return err, ""
	}

	err, rootFingerprint = rootBuild_stage3(
		ctx,
		rootBuild_stage3_request{
			Context: context,
			RootPath: rootPath,
			WorkspacePath: workspacePath,
			GardenPath: gardenPath,
		},
	)
	if err != nil {
		return err, ""
	}

	err, finalStemPath := rootBuild_stage4(
		ctx,
		rootBuild_stage4_request{
			Context: context,
			RootPath: rootPath,
			WorkspacePath: workspacePath,
			GardenPath: gardenPath,
		},
	)
	if err != nil {
		return err, ""
	}

	isUnstableRoot, err := fileExists(filepath.Join(finalStemPath, "dyd", "traits", "unstable"))
	if err != nil {
		return err, ""
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
			return err, ""
		}
	}

	if derivationFileExists {
		// fmt.Println("[trace] derivationFileExists " + derivationsPath)
		derivationsFingerprintFile := filepath.Join(derivationsPath, "dyd", "fingerprint")
		derivationsFingerprintBytes, err := ioutil.ReadFile(derivationsFingerprintFile)
		if err != nil {
			return err, ""
		}
		derivationsFingerprint := string(derivationsFingerprintBytes)

		stemBuildFingerprint = derivationsFingerprint

		// add the built fingerprint to the context
		context.FingerprintsMutex.Lock()		
		context.Fingerprints[rootPath] = derivationsFingerprint
		context.FingerprintsMutex.Unlock()

	} else {
		zlog.Info().
			Str("path", relRootPath).
			Msg("root build - building root")

		// otherwise run the root in a build env
		stemBuildPath, err := os.MkdirTemp("", "dryad-*")
		if err != nil {
			return err, ""
		}
		defer dydfs.RemoveAll(ctx, stemBuildPath)

		err, stemBuildFingerprint = rootBuild_stage5(
			ctx,
			rootBuild_stage5_request{
				RelRootPath: relRootPath,
				RootStemPath: finalStemPath,
				StemBuildPath: stemBuildPath,
				RootFingerprint: rootFingerprint,
			},
		)
		if err != nil {
			return err, ""
		}

		err, finalStemPath = rootBuild_stage6(
			ctx,
			rootBuild_stage6_request{
				RelRootPath: relRootPath,
				GardenPath: gardenPath,
				StemBuildPath: stemBuildPath,
			},
		)
		if err != nil {
			return err, ""
		}

		// add the built fingerprint to the context
		context.FingerprintsMutex.Lock()
		context.Fingerprints[rootPath] = stemBuildFingerprint
		context.FingerprintsMutex.Unlock()

		if !isUnstableRoot {
			// add the derivation link
			derivationsLinkPath, err := filepath.Rel(
				filepath.Dir(derivationsPath),
				finalStemPath,
			)
			if err != nil {
				return err, ""
			}
			err, _ = dydfs.RemoveAll(ctx, derivationsPath)
			if err != nil {
				return err, ""
			}
			err = os.Symlink(derivationsLinkPath, derivationsPath)
			if err != nil {
				return err, ""
			}
		}

		zlog.Info().
			Str("path", relRootPath).
			Msg("root build - done building root")
	}

	sproutPath := filepath.Join(gardenPath, "dyd", "sprouts", relRootPath)
	sproutParent := filepath.Dir(sproutPath)
	sproutHeapPath := filepath.Join(gardenPath, "dyd", "heap", "stems", stemBuildFingerprint)
	relSproutLink, err := filepath.Rel(
		sproutParent,
		sproutHeapPath,
	)
	if err != nil {
		return err, ""
	}

	// fmt.Println("[debug] building sprout parent")

	zlog.Debug().
		Str("path", sproutParent).
		Msg("root build - building sprout")
	err = dydfs.MkDir(sproutParent, fs.ModePerm)
	if err != nil {
		return err, ""
	}

	// fmt.Println("[debug] setting write permission on sprout parent")
	err = os.Chmod(sproutParent, 0o711)
	if err != nil {
		return err, ""
	}

	tmpSproutPath := sproutPath + ".tmp"
	zlog.Debug().
		Str("path", tmpSproutPath).
		Msg("root build - creating temporary sprout")
	// fmt.Println("[debug] adding temporary sprout link")
	err = os.Symlink(relSproutLink, tmpSproutPath)
	if err != nil {
		return err, ""
	}

	zlog.Debug().
		Str("tmpSproutPath", tmpSproutPath).
		Str("sproutPath", sproutPath).
		Msg("root build - renaming sprout")
	// fmt.Println("[debug] renaming sprout link", sproutPath)
	err = os.Rename(tmpSproutPath, sproutPath)
	if err != nil {
		return err, ""
	}

	// fmt.Println("[debug] setting read permissions on sprout parent")
	err = os.Chmod(sproutParent, 0o511)
	if err != nil {
		return err, ""
	}

	zlog.Info().
		Str("path", relRootPath).
		Msg("root build - done verifying root")

	return nil, stemBuildFingerprint
}

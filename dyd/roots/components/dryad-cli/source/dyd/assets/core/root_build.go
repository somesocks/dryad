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
	Garden *SafeGardenReference
	RootPath string
	JoinStdout bool
	JoinStderr bool
}

func rootBuild(ctx *task.ExecutionContext, req RootBuildRequest) (error, string) {
	var rootPath string = req.RootPath
	var gardenPath string = req.Garden.BasePath
	var err error

	relRootPath, err := filepath.Rel(
		filepath.Join(gardenPath, "dyd", "roots"),
		rootPath,
	)
	zlog.Debug().
		Str("gardenPath", gardenPath).
		Str("rootPath", rootPath).
		Str("relRootPath", relRootPath).
		Msg("RootBuild/relRootPath")
	if err != nil {
		return err, ""
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
			Garden: req.Garden,
			RootPath: rootPath,
			WorkspacePath: workspacePath,
			JoinStdout: req.JoinStdout,
			JoinStderr: req.JoinStderr,
		},
	)
	if err != nil {
		return err, ""
	}

	err, _ = rootBuild_stage2(
		ctx,
		rootBuild_stage2_request{
			RootPath: rootPath,
			WorkspacePath: workspacePath,
			GardenPath: gardenPath,
		},
	)
	if err != nil {
		return err, ""
	}

	err, rootFingerprint := rootBuild_stage3(
		ctx,
		rootBuild_stage3_request{
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
				Garden: req.Garden,
				RelRootPath: relRootPath,
				RootStemPath: finalStemPath,
				StemBuildPath: stemBuildPath,
				RootFingerprint: rootFingerprint,
				JoinStdout: req.JoinStdout,
				JoinStderr: req.JoinStderr,
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

	// create the sprout symlink
	zlog.Debug().
		Str("path", sproutPath).
		Str("target", relSproutLink).
		Msg("root build - building sprout symlink")
	err, _ = dydfs.Symlink(
		ctx,
		dydfs.SymlinkRequest{
			Target: relSproutLink,
			Path: sproutPath,
		},
	)
	if err != nil {
		return err, ""
	}

	zlog.Info().
		Str("path", relRootPath).
		Msg("root build - done verifying root")

	return nil, stemBuildFingerprint
}

var memoRootBuild = task.Memoize(rootBuild, "RootBuild") 

var RootBuild = func (ctx *task.ExecutionContext, req RootBuildRequest) (error, string) {
	var rootPath string = req.RootPath

	// sanitize the root path
	rootPath, err := RootPath(rootPath, "")
	zlog.Debug().
		Str("rootPath", rootPath).
		Str("gardenPath", req.Garden.BasePath).
		Msg("RootBuild/rootPath")
	if err != nil {
		return err, ""
	}

	req.RootPath = rootPath
	err, res := memoRootBuild(ctx, req)
	return err, res
}
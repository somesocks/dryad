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

type rootBuildRequest struct {
	Root *SafeRootReference
	JoinStdout bool
	JoinStderr bool
}

func rootBuild(ctx *task.ExecutionContext, req rootBuildRequest) (error, string) {
	var rootPath string = req.Root.BasePath
	var gardenPath string = req.Root.Roots.Garden.BasePath
	var err error

	relRootPath, err := filepath.Rel(
		req.Root.Roots.BasePath,
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
			Roots: req.Root.Roots,
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
			Garden: req.Root.Roots.Garden,
			RootPath: rootPath,
			WorkspacePath: workspacePath,
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
				Garden: req.Root.Roots.Garden,
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
				Garden: req.Root.Roots.Garden,
				RelRootPath: relRootPath,
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
			err, _ = dydfs.Symlink(
				ctx,
				dydfs.SymlinkRequest{
					Path: derivationsPath,
					Target: derivationsLinkPath,
				},
			)
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

var memoRootBuild = task.Memoize(
	rootBuild,
	func (req rootBuildRequest) any {
		return struct {
			Group string
			GardenPath string
			RootPath string
		}{
			Group: "RootBuild",
			GardenPath: req.Root.Roots.Garden.BasePath,
			RootPath: req.Root.BasePath,
		}
	},
)

var rootBuildWrapper = func (ctx *task.ExecutionContext, req rootBuildRequest) (error, string) {
	err, res := memoRootBuild(ctx, req)
	return err, res
}


type RootBuildRequest struct {
	JoinStdout bool
	JoinStderr bool
}

func (root *SafeRootReference) Build(ctx *task.ExecutionContext, req RootBuildRequest) (error, string) {
	err, res := rootBuildWrapper(
		ctx,
		rootBuildRequest{
			Root: root,
			JoinStdout: req.JoinStdout,
			JoinStderr: req.JoinStderr,	
		},
	)
	return err, res
}
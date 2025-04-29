package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"

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
	gardenRootPath := filepath.Join("dyd", "roots", relRootPath)

	if err != nil {
		return err, ""
	}

	zlog.Info().
		Str("path", gardenRootPath).
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

	err, rootStem := rootBuild_stage4(
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

	isUnstableRoot, err := fileExists(
		filepath.Join(rootStem.BasePath, "dyd", "traits", "unstable"))
	if err != nil {
		return err, ""
	}

	var stemBuildFingerprint string

	err, heap := req.Root.Roots.Garden.Heap().Resolve(ctx)
	if err != nil {
		return err, ""
	}

	err, heapDerivations := heap.Derivations().Resolve(ctx)
	if err != nil {
		return err, ""
	}

	unsafeDerivationRef := heapDerivations.Derivation(rootFingerprint)
	var derivationExists bool

	if !isUnstableRoot {
		err, derivationExists = unsafeDerivationRef.Exists(ctx)
		if err != nil {
			return err, ""
		}
	}

	if derivationExists {

		err, safeDerivationRef := unsafeDerivationRef.Resolve(ctx)
		if err != nil {
			return err, ""
		}

		derivationsFingerprint := filepath.Base(safeDerivationRef.Result.BasePath)

		stemBuildFingerprint = derivationsFingerprint

	} else {
		zlog.Info().
			Str("path", gardenRootPath).
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
				RootStemPath: rootStem.BasePath,
				StemBuildPath: stemBuildPath,
				RootFingerprint: rootFingerprint,
				JoinStdout: req.JoinStdout,
				JoinStderr: req.JoinStderr,
			},
		)
		if err != nil {
			return err, ""
		}

		err, _ = rootBuild_stage6(
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
			err, _ := heapDerivations.Add(
				ctx,
				rootFingerprint,
				stemBuildFingerprint,
			)
			if err != nil {
				return err, ""
			}
		}

		zlog.Info().
			Str("path", gardenRootPath).
			Msg("root build - done building root")
	}

	relSproutPath := filepath.Join("dyd", "sprouts", relRootPath) 
	sproutPath := filepath.Join(gardenPath, relSproutPath)
	sproutParent := filepath.Dir(sproutPath)
	sproutHeapPath := filepath.Join(gardenPath, "dyd", "heap", "stems", stemBuildFingerprint)
	relSproutLink, err := filepath.Rel(
		sproutParent,
		sproutHeapPath,
	)
	if err != nil {
		return err, ""
	}

	zlog.Info().
		Str("path", relSproutPath).
		Msg("root build - linking sprout")
	err, _ = dydfs.Mkdir2(
		ctx,
		dydfs.MkdirRequest{
			Path: sproutParent,
			Mode: 0o551,
			Recursive: true,
		},
	)
	if err != nil {
		zlog.Error().
			Str("sproutParent", sproutParent).
			Err(err).
			Msg("root build - building sprout parent")
		return err, ""
	}

	// create the sprout symlink
	zlog.Trace().
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
		zlog.Error().
			Str("sproutPath", sproutPath).
			Err(err).
			Msg("root build - building sprout symlink")
		return err, ""
	}

	zlog.Info().
		Str("path", gardenRootPath).
		Msg("root build - done verifying root")

	return nil, stemBuildFingerprint
}

var memoRootBuild = task.Memoize(
	rootBuild,
	func (ctx * task.ExecutionContext, req rootBuildRequest) (error, any) {
		var res = struct {
			Group string
			GardenPath string
			RootPath string
		}{
			Group: "RootBuild",
			GardenPath: req.Root.Roots.Garden.BasePath,
			RootPath: req.Root.BasePath,
		}
		return nil, res
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
package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"

	"errors"

	"os"
	"path/filepath"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

type rootBuildRequest struct {
	Root       *SafeRootReference
	JoinStdout bool
	JoinStderr bool
	LogStdout  struct {
		Path string
		Name string
	}
	LogStderr struct {
		Path string
		Name string
	}
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
			RootPath:      rootPath,
			WorkspacePath: workspacePath,
		},
	)
	if err != nil {
		return errors.New("error preparing root for build"), ""
	}

	err, _ = rootBuild_stage1(
		ctx,
		rootBuild_stage1_request{
			Roots:         req.Root.Roots,
			RootPath:      rootPath,
			WorkspacePath: workspacePath,
			JoinStdout:    req.JoinStdout,
			JoinStderr:    req.JoinStderr,
			LogStdout:     req.LogStdout,
			LogStderr:     req.LogStderr,
		},
	)
	if err != nil {
		return errors.New("error resolving root dependencies"), ""
	}

	err, _ = rootBuild_stage2(
		ctx,
		rootBuild_stage2_request{
			RootPath:      rootPath,
			WorkspacePath: workspacePath,
			GardenPath:    gardenPath,
		},
	)
	if err != nil {
		return errors.New("error preparing root execution path"), ""
	}

	err, rootFingerprint := rootBuild_stage3(
		ctx,
		rootBuild_stage3_request{
			RootPath:      rootPath,
			WorkspacePath: workspacePath,
			GardenPath:    gardenPath,
		},
	)
	if err != nil {
		return errors.New("error generating root fingerprint"), ""
	}

	err, rootStem := rootBuild_stage4(
		ctx,
		rootBuild_stage4_request{
			Garden:        req.Root.Roots.Garden,
			RootPath:      rootPath,
			WorkspacePath: workspacePath,
		},
	)
	if err != nil {
		return errors.New("error packing root into heap"), ""
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
				Garden:          req.Root.Roots.Garden,
				RelRootPath:     relRootPath,
				RootStemPath:    rootStem.BasePath,
				StemBuildPath:   stemBuildPath,
				RootFingerprint: rootFingerprint,
				JoinStdout:      req.JoinStdout,
				JoinStderr:      req.JoinStderr,
				LogStdout:       req.LogStdout,
				LogStderr:       req.LogStderr,
			},
		)
		if err != nil {
			return errors.New("error executing root to build stem"), ""
		}

		err, _ = rootBuild_stage6(
			ctx,
			rootBuild_stage6_request{
				Garden:        req.Root.Roots.Garden,
				RelRootPath:   relRootPath,
				StemBuildPath: stemBuildPath,
			},
		)
		if err != nil {
			return errors.New("error packing stem into heap"), ""
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

	// build and publish a sprout package for this root
	sproutBuildPath, err := os.MkdirTemp("", "dryad-*")
	if err != nil {
		return err, ""
	}
	defer dydfs.RemoveAll(ctx, sproutBuildPath)

	err = SproutInit(sproutBuildPath)
	if err != nil {
		return errors.New("error preparing sprout workspace"), ""
	}

	rootTraitsPath := filepath.Join(rootPath, "dyd", "traits")
	rootTraitsExists, err := fileExists(rootTraitsPath)
	if err != nil {
		return err, ""
	}
	if rootTraitsExists {
		err = rootDevelop_copyDir(
			task.SERIAL_CONTEXT,
			rootTraitsPath,
			filepath.Join(sproutBuildPath, "dyd", "traits"),
			rootDevelopCopyOptions{ApplyIgnore: false},
		)
		if err != nil {
			return errors.New("error copying root traits into sprout"), ""
		}
	}

	builtStemPath := filepath.Join(gardenPath, "dyd", "heap", "stems", stemBuildFingerprint)
	sproutDependenciesPath := filepath.Join(sproutBuildPath, "dyd", "dependencies")

	err = os.Symlink(
		builtStemPath,
		filepath.Join(sproutDependenciesPath, "stem"),
	)
	if err != nil {
		return errors.New("error linking stem dependency for sprout"), ""
	}

	err = sproutRequirementsPrepare(sproutBuildPath)
	if err != nil {
		return errors.New("error preparing sprout requirements"), ""
	}

	err, _ = sproutFinalize(ctx, sproutBuildPath)
	if err != nil {
		return errors.New("error finalizing sprout package"), ""
	}

	err, heapSprouts := heap.Sprouts().Resolve(ctx)
	if err != nil {
		return err, ""
	}

	err, heapSprout := heapSprouts.AddSprout(
		ctx,
		HeapAddSproutRequest{
			SproutPath: sproutBuildPath,
		},
	)
	if err != nil {
		return errors.New("error packing sprout into heap"), ""
	}

	relSproutPath := filepath.Join("dyd", "sprouts", relRootPath)
	sproutPath := filepath.Join(gardenPath, relSproutPath)
	sproutParent := filepath.Dir(sproutPath)
	sproutsPath := filepath.Join(gardenPath, "dyd", "sprouts")
	sproutHeapPath := heapSprout.BasePath
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

	relSproutParentPath := filepath.Dir(relRootPath)
	if relSproutParentPath != "." {
		var currentPath = sproutsPath
		for _, part := range strings.Split(relSproutParentPath, string(filepath.Separator)) {
			if part == "" || part == "." {
				continue
			}

			currentPath = filepath.Join(currentPath, part)
			currentInfo, currentErr := os.Lstat(currentPath)
			if currentErr != nil {
				if os.IsNotExist(currentErr) {
					continue
				}
				return currentErr, ""
			}

			if !currentInfo.IsDir() {
				currentErr, _ = dydfs.Remove(ctx, currentPath)
				if currentErr != nil {
					return currentErr, ""
				}
			}
		}
	}

	err, _ = dydfs.Mkdir2(
		ctx,
		dydfs.MkdirRequest{
			Path:      sproutParent,
			Mode:      0o551,
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
			Path:   sproutPath,
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

var rootBuild2 = task.OnFailure(
	rootBuild,
	func(ctx *task.ExecutionContext, args task.Tuple2[rootBuildRequest, error]) (error, any) {
		var req = args.A
		var err = args.B

		zlog.
			Error().
			Err(err).
			Str("path", req.Root.BasePath).
			Msg("error while building root")

		return nil, nil
	},
)

var memoRootBuild = task.Memoize(
	rootBuild2,
	func(ctx *task.ExecutionContext, req rootBuildRequest) (error, any) {
		var res = struct {
			Group      string
			GardenPath string
			RootPath   string
		}{
			Group:      "RootBuild",
			GardenPath: req.Root.Roots.Garden.BasePath,
			RootPath:   req.Root.BasePath,
		}
		return nil, res
	},
)

var rootBuildWrapper = func(ctx *task.ExecutionContext, req rootBuildRequest) (error, string) {
	err, res := memoRootBuild(ctx, req)
	return err, res
}

type RootBuildRequest struct {
	JoinStdout bool
	JoinStderr bool
	LogStdout  struct {
		Path string
		Name string
	}
	LogStderr struct {
		Path string
		Name string
	}
}

func (root *SafeRootReference) Build(ctx *task.ExecutionContext, req RootBuildRequest) (error, string) {
	err, res := rootBuildWrapper(
		ctx,
		rootBuildRequest{
			Root:       root,
			JoinStdout: req.JoinStdout,
			JoinStderr: req.JoinStderr,
			LogStdout:  req.LogStdout,
			LogStderr:  req.LogStderr,
		},
	)
	return err, res
}

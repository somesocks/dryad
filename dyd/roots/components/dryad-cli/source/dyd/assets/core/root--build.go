package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"

	"fmt"

	"os"
	"path/filepath"
	"sort"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

type rootBuildRequest struct {
	Root              *SafeRootReference
	VariantDescriptor string
	JoinStdout        bool
	JoinStderr        bool
	LogStdout         struct {
		Path string
		Name string
	}
	LogStderr struct {
		Path string
		Name string
	}
}

type rootMaterializeSproutRequest struct {
	Root          *SafeRootReference
	StemByVariant map[string]string
}

func rootMaterializeSprout(ctx *task.ExecutionContext, req rootMaterializeSproutRequest) (error, string) {
	rootPath := req.Root.BasePath
	gardenPath := req.Root.Roots.Garden.BasePath

	relRootPath, err := filepath.Rel(
		req.Root.Roots.BasePath,
		rootPath,
	)
	if err != nil {
		return err, ""
	}

	if len(req.StemByVariant) == 0 {
		return fmt.Errorf("no stem variants provided for sprout materialization: %s", rootPath), ""
	}

	sproutBuildPath, err := os.MkdirTemp("", "dryad-*")
	if err != nil {
		return err, ""
	}
	defer dydfs.RemoveAll(ctx, sproutBuildPath)

	err = SproutInit(sproutBuildPath)
	if err != nil {
		return fmt.Errorf("error preparing sprout workspace: %w", err), ""
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
			return fmt.Errorf("error copying root traits into sprout: %w", err), ""
		}
	}

	sproutDependenciesPath := filepath.Join(sproutBuildPath, "dyd", "dependencies")
	descriptors := make([]string, 0, len(req.StemByVariant))
	for descriptor := range req.StemByVariant {
		descriptors = append(descriptors, descriptor)
	}
	sort.Strings(descriptors)

	for _, descriptor := range descriptors {
		stemFingerprint := req.StemByVariant[descriptor]
		if strings.TrimSpace(stemFingerprint) == "" {
			return fmt.Errorf("empty stem fingerprint for variant descriptor: %s", descriptor), ""
		}

		dependencyName := "stem"
		if descriptor != "" {
			dependencyName = dependencyName + RootRequirementSelectorSeparator + descriptor
		}

		builtStemPath := filepath.Join(gardenPath, "dyd", "heap", "stems", stemFingerprint)
		err = os.Symlink(
			builtStemPath,
			filepath.Join(sproutDependenciesPath, dependencyName),
		)
		if err != nil {
			return fmt.Errorf("error linking stem dependency for sprout: %w", err), ""
		}
	}

	err = sproutRequirementsPrepare(sproutBuildPath)
	if err != nil {
		return fmt.Errorf("error preparing sprout requirements: %w", err), ""
	}

	err, sproutFingerprint := sproutFinalize(ctx, sproutBuildPath)
	if err != nil {
		return fmt.Errorf("error finalizing sprout package: %w", err), ""
	}

	err, heap := req.Root.Roots.Garden.Heap().Resolve(ctx)
	if err != nil {
		return err, ""
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
		return fmt.Errorf("error packing sprout into heap: %w", err), ""
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
		currentPath := sproutsPath
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

	return nil, sproutFingerprint
}

func rootBuildStem(ctx *task.ExecutionContext, req rootBuildRequest) (error, string) {
	rootPath := req.Root.BasePath
	gardenPath := req.Root.Roots.Garden.BasePath
	variantLabel := rootBuildLogVariantLabel(req.VariantDescriptor)

	relRootPath, err := filepath.Rel(
		req.Root.Roots.BasePath,
		rootPath,
	)
	if err != nil {
		return err, ""
	}
	gardenRootPath := filepath.Join("dyd", "roots", relRootPath)

	zlog.Info().
		Str("path", gardenRootPath).
		Str("variant", variantLabel).
		Msg("root build - verifying root")

	workspacePath, err := os.MkdirTemp("", "dryad-*")
	if err != nil {
		return err, ""
	}
	defer dydfs.RemoveAll(ctx, workspacePath)

	err, _ = rootBuild_stage0(
		ctx,
		rootBuild_stage0_request{
			RootPath:          rootPath,
			WorkspacePath:     workspacePath,
			VariantDescriptor: req.VariantDescriptor,
		},
	)
	if err != nil {
		return fmt.Errorf("error preparing root for build: %w", err), ""
	}

	err, _ = rootBuild_stage1(
		ctx,
		rootBuild_stage1_request{
			Roots:             req.Root.Roots,
			RootPath:          rootPath,
			WorkspacePath:     workspacePath,
			VariantDescriptor: req.VariantDescriptor,
			JoinStdout:        req.JoinStdout,
			JoinStderr:        req.JoinStderr,
			LogStdout:         req.LogStdout,
			LogStderr:         req.LogStderr,
		},
	)
	if err != nil {
		return fmt.Errorf("error resolving root dependencies: %w", err), ""
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
		return fmt.Errorf("error preparing root execution path: %w", err), ""
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
		return fmt.Errorf("error generating root fingerprint: %w", err), ""
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
		return fmt.Errorf("error packing root into heap: %w", err), ""
	}

	isUnstableRoot, err := fileExists(
		filepath.Join(rootStem.BasePath, "dyd", "traits", "unstable"),
	)
	if err != nil {
		return err, ""
	}

	err, heap := req.Root.Roots.Garden.Heap().Resolve(ctx)
	if err != nil {
		return err, ""
	}

	err, heapDerivations := heap.Derivations().Resolve(ctx)
	if err != nil {
		return err, ""
	}

	unsafeDerivationRef := heapDerivations.Derivation(rootFingerprint)
	derivationExists := false

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
		return nil, derivationsFingerprint
	}

	zlog.Info().
		Str("path", gardenRootPath).
		Str("variant", variantLabel).
		Msg("root build - building root")

	stemBuildPath, err := os.MkdirTemp("", "dryad-*")
	if err != nil {
		return err, ""
	}
	defer dydfs.RemoveAll(ctx, stemBuildPath)

	err, stemBuildFingerprint := rootBuild_stage5(
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
		return fmt.Errorf("error executing root to build stem: %w", err), ""
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
		return fmt.Errorf("error packing stem into heap: %w", err), ""
	}

	if !isUnstableRoot {
		err, _ = heapDerivations.Add(
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
		Str("variant", variantLabel).
		Msg("root build - done building root")

	return nil, stemBuildFingerprint
}

var rootBuildStem2 = task.OnFailure(
	rootBuildStem,
	func(ctx *task.ExecutionContext, args task.Tuple2[rootBuildRequest, error]) (error, any) {
		req := args.A
		err := args.B

		zlog.
			Error().
			Err(err).
			Str("path", req.Root.BasePath).
			Str("variant", rootBuildLogVariantLabel(req.VariantDescriptor)).
			Msg("error while building root")

		return nil, nil
	},
)

var memoRootBuildStem = task.Memoize(
	rootBuildStem2,
	func(ctx *task.ExecutionContext, req rootBuildRequest) (error, any) {
		res := struct {
			Group             string
			GardenPath        string
			RootPath          string
			VariantDescriptor string
		}{
			Group:             "RootBuildStem",
			GardenPath:        req.Root.Roots.Garden.BasePath,
			RootPath:          req.Root.BasePath,
			VariantDescriptor: req.VariantDescriptor,
		}
		return nil, res
	},
)

var rootBuildStemWrapper = func(ctx *task.ExecutionContext, req rootBuildRequest) (error, string) {
	err, res := memoRootBuildStem(ctx, req)
	return err, res
}

type RootBuildRequest struct {
	VariantDescriptor string
	JoinStdout        bool
	JoinStderr        bool
	LogStdout         struct {
		Path string
		Name string
	}
	LogStderr struct {
		Path string
		Name string
	}
}

type RootBuildStemRequest = RootBuildRequest

type RootBuildSproutRequest = RootBuildRequest

func rootBuildLogVariantLabel(variantDescriptor string) string {
	if variantDescriptor == "" {
		return "default"
	}

	return variantDescriptor
}

func normalizeRootBuildVariantDescriptor(raw string) (error, string) {
	err, variantContext := RootVariantContextFromFilesystem(raw)
	if err != nil {
		return err, ""
	}

	err, variantDescriptor := variantContext.Filesystem()
	if err != nil {
		return err, ""
	}

	return nil, variantDescriptor
}

func (root *SafeRootReference) BuildStem(ctx *task.ExecutionContext, req RootBuildStemRequest) (error, string) {
	err, variantDescriptor := normalizeRootBuildVariantDescriptor(req.VariantDescriptor)
	if err != nil {
		return err, ""
	}

	err, res := rootBuildStemWrapper(
		ctx,
		rootBuildRequest{
			Root:              root,
			VariantDescriptor: variantDescriptor,
			JoinStdout:        req.JoinStdout,
			JoinStderr:        req.JoinStderr,
			LogStdout:         req.LogStdout,
			LogStderr:         req.LogStderr,
		},
	)
	return err, res
}

func (root *SafeRootReference) BuildSprout(ctx *task.ExecutionContext, req RootBuildSproutRequest) (error, string) {
	err, variantDescriptor := normalizeRootBuildVariantDescriptor(req.VariantDescriptor)
	if err != nil {
		return err, ""
	}

	err, variantSelector := variantDescriptorParseFilesystem(variantDescriptor)
	if err != nil {
		return err, ""
	}

	err, variants := root.ResolveBuildVariants(
		ctx,
		RootResolveBuildVariantsRequest{
			Selector:                variantSelector,
			IgnoreUnknownDimensions: true,
		},
	)
	if err != nil {
		return err, ""
	}

	type rootBuildVariantResult struct {
		Descriptor      string
		StemFingerprint string
	}

	buildVariant := func(ctx *task.ExecutionContext, variant VariantDescriptor) (error, rootBuildVariantResult) {
		err, concreteDescriptor := variantDescriptorEncodeFilesystem(variant)
		if err != nil {
			return err, rootBuildVariantResult{}
		}

		err, stemFingerprint := root.BuildStem(
			ctx,
			RootBuildStemRequest{
				VariantDescriptor: concreteDescriptor,
				JoinStdout:        req.JoinStdout,
				JoinStderr:        req.JoinStderr,
				LogStdout:         req.LogStdout,
				LogStderr:         req.LogStderr,
			},
		)
		if err != nil {
			return err, rootBuildVariantResult{}
		}

		return nil, rootBuildVariantResult{
			Descriptor:      concreteDescriptor,
			StemFingerprint: stemFingerprint,
		}
	}

	err, builtVariants := task.ParallelMap(buildVariant)(ctx, variants)
	if err != nil {
		return err, ""
	}

	stemByVariant := map[string]string{}
	for _, builtVariant := range builtVariants {
		if _, exists := stemByVariant[builtVariant.Descriptor]; exists {
			return fmt.Errorf("duplicate root build variant descriptor: %s", builtVariant.Descriptor), ""
		}

		stemByVariant[builtVariant.Descriptor] = builtVariant.StemFingerprint
	}

	err, sproutFingerprint := rootMaterializeSprout(
		ctx,
		rootMaterializeSproutRequest{
			Root:          root,
			StemByVariant: stemByVariant,
		},
	)
	if err != nil {
		return err, ""
	}

	relRootPath, err := filepath.Rel(root.Roots.BasePath, root.BasePath)
	if err != nil {
		return err, ""
	}
	gardenRootPath := filepath.Join("dyd", "roots", relRootPath)

	zlog.Info().
		Str("path", gardenRootPath).
		Msg("root build - done verifying root")

	return nil, sproutFingerprint
}

func (root *SafeRootReference) Build(ctx *task.ExecutionContext, req RootBuildRequest) (error, string) {
	return root.BuildStem(ctx, req)
}

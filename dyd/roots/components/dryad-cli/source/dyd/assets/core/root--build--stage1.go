package core

import (
	// dydfs "dryad/filesystem"
	"dryad/internal/os"
	"dryad/task"
	"fmt"

	// "io/fs"
	// "io/ioutil"
	"dryad/internal/filepath"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage1_request struct {
	Roots                    *SafeRootsReference
	RootPath                 string
	WorkspacePath            string
	VariantDescriptor        string
	SelectedRequirementsPath string
	JoinStdout               bool
	JoinStderr               bool
	LogStdout                struct {
		Path string
		Name string
	}
	LogStderr struct {
		Path string
		Name string
	}
}

func rootBuild_stage1DependencyName(
	requirementName string,
	target RootRequirementResolvedTarget,
	totalTargets int,
) (error, string) {
	dependencyName := requirementName
	if totalTargets > 1 || target.ForceVariantSuffix {
		err, descriptorSuffix := variantDescriptorEncodeFilesystem(target.VariantDescriptor)
		if err != nil {
			return err, ""
		}
		if descriptorSuffix != "" {
			dependencyName = dependencyName + RootRequirementSelectorSeparator + descriptorSuffix
		}
	}

	return nil, dependencyName
}

// stage 1 - walk through the root dependencies, build them if necessary,
// and add the fingerprint as a dependency
var rootBuild_stage1 func(ctx *task.ExecutionContext, req rootBuild_stage1_request) (error, []*RootBuildResult)

func init() {

	// the initialization for rootBuild_stage1 has to be deferred in an init block,
	// in order to avoid an init cycle with RootBuild
	type rootBuild_stage1_buildDependencyRequest struct {
		BaseRequest                 rootBuild_stage1_request
		DependencyName              string
		DependencyRoot              *SafeRootReference
		DependencyPath              string
		DependencyVariantDescriptor string
		FileTargetSpec              *RootRequirementTargetSpec
		JoinStdout                  bool
		JoinStderr                  bool
		LogStdout                   struct {
			Path string
			Name string
		}
		LogStderr struct {
			Path string
			Name string
		}
	}

	var rootBuild_stage1_prepReq = func(ctx *task.ExecutionContext, req rootBuild_stage1_request) (error, rootBuild_stage1_request) {
		zlog.Trace().
			Msg("RootBuild/stage1")

		if err := os.Mkdir(filepath.Join(req.WorkspacePath, "dyd", "requirements"), os.ModePerm); err != nil {
			return err, req
		}

		if err := os.Mkdir(filepath.Join(req.WorkspacePath, "dyd", "path"), os.ModePerm); err != nil {
			return err, req
		}

		return nil, req
	}

	var rootBuild_stage1_generateRequests = func(
		ctx *task.ExecutionContext,
		req rootBuild_stage1_request,
	) (error, []rootBuild_stage1_buildDependencyRequest) {
		var buildDependencyRequests []rootBuild_stage1_buildDependencyRequest
		boundNames := map[string]struct{}{}

		bindName := func(name string) error {
			if _, exists := boundNames[name]; exists {
				return fmt.Errorf("duplicate materialized requirement name: %s", name)
			}
			boundNames[name] = struct{}{}
			return nil
		}

		materializeEnvRequirement := func(requirementName string, targetSpec *RootRequirementTargetSpec) error {
			err, injectName := rootRequirementCanonicalEnvName(requirementName)
			if err != nil {
				return err
			}
			if err := bindName(injectName); err != nil {
				return err
			}

			envValue, exists := os.LookupEnv(targetSpec.EnvName)
			if !exists {
				return fmt.Errorf("missing env requirement %s: host env %s is not set", injectName, targetSpec.EnvName)
			}

			err, envFingerprint := rootRequirementEnvValueFingerprint(envValue)
			if err != nil {
				return err
			}
			if targetSpec.EnvFingerprint != "" && targetSpec.EnvFingerprint != envFingerprint {
				return fmt.Errorf("env requirement %s fingerprint mismatch for host env %s", injectName, targetSpec.EnvName)
			}

			return os.WriteFile(
				filepath.Join(req.WorkspacePath, "dyd", "requirements", injectName),
				[]byte(rootRequirementEnvTargetString(targetSpec.EnvName, envFingerprint)),
				0o511,
			)
		}

		if req.SelectedRequirementsPath == "" {
			return nil, buildDependencyRequests
		}

		rootRef := SafeRootReference{
			BasePath: req.RootPath,
			Roots:    req.Roots,
		}

		err, parentVariantContext := RootVariantContextFromFilesystem(req.VariantDescriptor)
		if err != nil {
			return err, buildDependencyRequests
		}

		requirementsRef := SafeRootRequirementsReference{
			BasePath: req.SelectedRequirementsPath,
			Root:     &rootRef,
		}

		err = requirementsRef.Walk(task.SERIAL_CONTEXT, RootRequirementsWalkRequest{
			OnMatch: func(ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any) {
				err, requirementName, condition := rootRequirementParseName(filepath.Base(requirement.BasePath))
				if err != nil {
					return err, nil
				}

				err, shouldInclude := rootRequirementConditionMatches(
					parentVariantContext.Descriptor,
					condition,
				)
				if err != nil {
					return err, nil
				}
				if !shouldInclude {
					return nil, nil
				}

				err, targetSpec := requirement.TargetSpec(ctx)
				if err != nil {
					return err, nil
				}
				if rootRequirementTargetKind(targetSpec.Kind) == RootRequirementTargetKindEnv {
					return materializeEnvRequirement(requirementName, targetSpec), nil
				}
				if rootRequirementTargetKind(targetSpec.Kind) == RootRequirementTargetKindFile {
					if err := bindName(requirementName); err != nil {
						return err, nil
					}
					buildDependencyRequests = append(buildDependencyRequests, rootBuild_stage1_buildDependencyRequest{
						BaseRequest:    req,
						DependencyName: requirementName,
						DependencyPath: targetSpec.FileSourcePath,
						FileTargetSpec: targetSpec,
					})
					return nil, nil
				}

				err, targets := requirement.ResolveTargets(ctx, RootRequirementResolveTargetsRequest{
					ParentVariant: parentVariantContext.Descriptor,
				})
				if err != nil {
					return err, nil
				}

				for _, target := range targets {
					err, dependencyName := rootBuild_stage1DependencyName(requirementName, target, len(targets))
					if err != nil {
						return err, nil
					}
					if err := bindName(dependencyName); err != nil {
						return err, nil
					}

					err, dependencyVariantDescriptor := variantDescriptorEncodeFilesystem(target.VariantDescriptor)
					if err != nil {
						return err, nil
					}

					buildDependencyRequests = append(buildDependencyRequests, rootBuild_stage1_buildDependencyRequest{
						BaseRequest:                 req,
						DependencyName:              dependencyName,
						DependencyRoot:              target.Root,
						DependencyPath:              target.Root.BasePath,
						DependencyVariantDescriptor: dependencyVariantDescriptor,
						JoinStdout:                  req.JoinStdout,
						JoinStderr:                  req.JoinStderr,
						LogStdout:                   req.LogStdout,
						LogStderr:                   req.LogStderr,
					})
				}

				return nil, nil
			},
		})
		if err != nil {
			return err, buildDependencyRequests
		}

		return nil, buildDependencyRequests
	}

	var rootBuild_stage1_buildDependency = func(ctx *task.ExecutionContext, req rootBuild_stage1_buildDependencyRequest) (error, *RootBuildResult) {
		if req.FileTargetSpec != nil {
			err, fileStem := RootRequirementFileBuildStem(ctx, RootRequirementFileBuildStemRequest{
				Garden:          req.BaseRequest.Roots.Garden,
				SourcePath:      req.FileTargetSpec.FileSourcePath,
				DestinationAs:   req.FileTargetSpec.FileDestinationAs,
				DestinationInto: req.FileTargetSpec.FileDestinationInto,
				Optional:        req.FileTargetSpec.FileOptional,
				Unpack:          req.FileTargetSpec.FileUnpack,
			})
			if err != nil {
				return err, nil
			}
			if fileStem == nil {
				return fmt.Errorf("missing file dependency build result: %s", req.DependencyPath), nil
			}
			if req.FileTargetSpec.FileFingerprint != "" && req.FileTargetSpec.FileFingerprint != fileStem.Fingerprint {
				return fmt.Errorf("file requirement %s fingerprint mismatch for %s", req.DependencyName, req.FileTargetSpec.FileSourcePath), nil
			}

			err = rootBuild_linkDependency(rootBuild_linkDependencyRequest{
				WorkspacePath:         req.BaseRequest.WorkspacePath,
				DependencyName:        req.DependencyName,
				DependencyHeapPath:    fileStem.BasePath,
				DependencyFingerprint: fileStem.Fingerprint,
			})
			if err != nil {
				return err, nil
			}

			return nil, &RootBuildResult{
				SourceFingerprint: fileStem.Fingerprint,
				ResultFingerprint: fileStem.Fingerprint,
			}
		}

		if req.DependencyRoot == nil {
			return fmt.Errorf("missing dependency root: %s", req.DependencyPath), nil
		}

		err, dependencyBuildResult := req.DependencyRoot.BuildStem(
			ctx,
			RootBuildStemRequest{
				VariantDescriptor: req.DependencyVariantDescriptor,
				JoinStdout:        req.JoinStdout,
				JoinStderr:        req.JoinStderr,
				LogStdout:         req.LogStdout,
				LogStderr:         req.LogStderr,
			},
		)
		if err != nil {
			return err, nil
		}
		if dependencyBuildResult == nil {
			return fmt.Errorf("missing dependency build result: %s", req.DependencyPath), nil
		}

		err, dependencyHeapPath := heapStemsFingerprintPath(
			ctx,
			req.BaseRequest.Roots.Garden,
			filepath.Join(req.BaseRequest.Roots.Garden.BasePath, "dyd", "heap", "stems"),
			dependencyBuildResult.ResultFingerprint,
		)
		if err != nil {
			return err, nil
		}

		err = rootBuild_linkDependency(rootBuild_linkDependencyRequest{
			WorkspacePath:         req.BaseRequest.WorkspacePath,
			DependencyName:        req.DependencyName,
			DependencyHeapPath:    dependencyHeapPath,
			DependencyFingerprint: dependencyBuildResult.ResultFingerprint,
		})
		if err != nil {
			return err, nil
		}

		return nil, dependencyBuildResult
	}

	var rootBuild_stage1_buildDependencies = task.ParallelMap(rootBuild_stage1_buildDependency)

	var rootBuild_stage1_processResults = func(ctx *task.ExecutionContext, req []*RootBuildResult) (error, []*RootBuildResult) {
		return nil, req
	}

	rootBuild_stage1 = task.Series4(
		rootBuild_stage1_prepReq,
		rootBuild_stage1_generateRequests,
		rootBuild_stage1_buildDependencies,
		rootBuild_stage1_processResults,
	)

}

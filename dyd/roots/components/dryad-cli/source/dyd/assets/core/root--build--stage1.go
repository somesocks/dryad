package core

import (
	// dydfs "dryad/filesystem"
	"dryad/task"

	// "io/fs"
	// "io/ioutil"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage1_request struct {
	Roots *SafeRootsReference
	RootPath string
	WorkspacePath string
	JoinStdout bool
	JoinStderr bool
	LogStdout struct {
		Path string
		Name string
	}
	LogStderr struct {
		Path string
		Name string
	}
}

// stage 1 - walk through the root dependencies, build them if necessary,
// and add the fingerprint as a dependency
var rootBuild_stage1 func (ctx *task.ExecutionContext, req rootBuild_stage1_request) (error, any)


func init () {

	// the initialization for rootBuild_stage1 has to be deferred in an init block,
	// in order to avoid an init cycle with RootBuild
	type rootBuild_stage1_buildDependencyRequest struct {
		Roots *SafeRootsReference
		BaseRequest rootBuild_stage1_request
		DependencyName string
		DependencyPath string
		JoinStdout bool
		JoinStderr bool
		LogStdout struct {
			Path string
			Name string
		}
		LogStderr struct {
			Path string
			Name string
		}
	}
	
	var rootBuild_stage1_prepReq = func (ctx *task.ExecutionContext, req rootBuild_stage1_request) (error, rootBuild_stage1_request) {
		zlog.Trace().
			Msg("RootBuild/stage1")
	
		return nil, req
	}
	
	var rootBuild_stage1_generateRequests = func (
		ctx *task.ExecutionContext, 
		req rootBuild_stage1_request,
	) (error, []rootBuild_stage1_buildDependencyRequest) {
		var buildDependencyRequests []rootBuild_stage1_buildDependencyRequest

		rootRef := SafeRootReference{
			BasePath: req.RootPath,
			Roots: req.Roots,
		}

		err, requirementsRef := rootRef.Requirements().Resolve(ctx)
		if err != nil {
			return err, buildDependencyRequests
		}

		err = requirementsRef.Walk(task.SERIAL_CONTEXT, RootRequirementsWalkRequest{
			OnMatch: func (ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any) {
				err, target := requirement.Target(ctx)
				if err != nil {
					return err, nil
				}

				buildDependencyRequests = append(buildDependencyRequests, rootBuild_stage1_buildDependencyRequest{
					Roots: req.Roots,
					BaseRequest: req,
					DependencyName: filepath.Base(requirement.BasePath),
					DependencyPath: target.BasePath,
					JoinStdout: req.JoinStdout,
					JoinStderr: req.JoinStderr,
					LogStdout: req.LogStdout,
					LogStderr: req.LogStderr,
				})	

				return nil, nil
			},
		});
		if err != nil {
			return err, buildDependencyRequests
		}
	
		return nil, buildDependencyRequests
	}
	
	var rootBuild_stage1_buildDependency = func (ctx *task.ExecutionContext, req rootBuild_stage1_buildDependencyRequest) (error, string) {
	
		var unsafeDepReference = UnsafeRootReference{
			Roots: req.Roots,
			BasePath: req.DependencyPath,
		}

		var safeDepReference SafeRootReference
		var err error

		// verify that root path is valid for dependency
		err, safeDepReference = unsafeDepReference.Resolve(ctx)
		if err != nil {
			return err, ""
		}
	
		err, dependencyFingerprint := safeDepReference.Build(
			ctx,
			RootBuildRequest{
				JoinStdout: req.JoinStdout,
				JoinStderr: req.JoinStderr,
				LogStdout: req.LogStdout,
				LogStderr: req.LogStderr,
			},
		)
		if err != nil {
			return err, ""
		}
	

		
		dependencyHeapPath := filepath.Join(req.BaseRequest.Roots.Garden.BasePath, "dyd", "heap", "stems", dependencyFingerprint)
	
		dependencyName := req.DependencyName
	
		targetDepPath := filepath.Join(req.BaseRequest.WorkspacePath, "dyd", "dependencies", dependencyName)
	
		err = os.Symlink(dependencyHeapPath, targetDepPath)
	
		if err != nil {
			return err, ""
		}
	
		return nil, ""
	}
	
	var rootBuild_stage1_buildDependencies = task.ParallelMap(rootBuild_stage1_buildDependency)
	
	var rootBuild_stage1_processResults = func (ctx *task.ExecutionContext, req []string) (error, any) {
		return nil, nil
	}
	
	rootBuild_stage1 = task.Series4(
		rootBuild_stage1_prepReq,
		rootBuild_stage1_generateRequests,
		rootBuild_stage1_buildDependencies,
		rootBuild_stage1_processResults,
	)
	
}
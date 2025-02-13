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
	Garden *SafeGardenReference
	RootPath string
	WorkspacePath string
	JoinStdout bool
	JoinStderr bool
}

// stage 1 - walk through the root dependencies, build them if necessary,
// and add the fingerprint as a dependency
var rootBuild_stage1 func (ctx *task.ExecutionContext, req rootBuild_stage1_request) (error, any)


func init () {

	// the initialization for rootBuild_stage1 has to be deferred in an init block,
	// in order to avoid an init cycle with RootBuild
	type rootBuild_stage1_buildDependencyRequest struct {
		Garden *SafeGardenReference
		BaseRequest rootBuild_stage1_request
		DependencyPath string
		JoinStdout bool
		JoinStderr bool
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
		var rootsPath string = filepath.Join(req.RootPath, "dyd", "requirements")
		var buildDependencyRequests []rootBuild_stage1_buildDependencyRequest
	
		dependencies, err := filepath.Glob(filepath.Join(rootsPath, "*"))
		if err != nil {
			return err, buildDependencyRequests
		}
	
		for _, dependencyPath := range dependencies {
			buildDependencyRequests = append(buildDependencyRequests, rootBuild_stage1_buildDependencyRequest{
				Garden: req.Garden,
				BaseRequest: req,
				DependencyPath: dependencyPath,
				JoinStdout: req.JoinStdout,
				JoinStderr: req.JoinStderr,
			})
		}
	
		return nil, buildDependencyRequests
	}
	
	var rootBuild_stage1_buildDependency = func (ctx *task.ExecutionContext, req rootBuild_stage1_buildDependencyRequest) (error, string) {
	
		var unsafeDepReference = UnsafeRootReference{
			Garden: req.Garden,
			BasePath: req.DependencyPath,
		}

		var safeDepReference SafeRootReference
		var err error

		// verify that root path is valid for dependency
		err, safeDepReference = unsafeDepReference.Resolve(ctx, nil)
		if err != nil {
			return err, ""
		}
	
		err, dependencyFingerprint := RootBuild(
			ctx,
			RootBuildRequest{
				Root: &safeDepReference,
				JoinStdout: req.JoinStdout,
				JoinStderr: req.JoinStderr,
			},
		)
		if err != nil {
			return err, ""
		}
	
		dependencyHeapPath := filepath.Join(req.BaseRequest.Garden.BasePath, "dyd", "heap", "stems", dependencyFingerprint)
	
		dependencyName := filepath.Base(req.DependencyPath)
	
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
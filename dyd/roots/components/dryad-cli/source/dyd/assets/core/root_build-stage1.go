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
	Context BuildContext
	RootPath string
	WorkspacePath string
	GardenPath string
}

// stage 1 - walk through the root dependencies, build them if necessary,
// and add the fingerprint as a dependency
var rootBuild_stage1 func (ctx *task.ExecutionContext, req rootBuild_stage1_request) (error, any)


func init () {

	// the initialization for rootBuild_stage1 has to be deferred to init init,
	// in order to avoid an init cycle with RootBuild
	type rootBuild_stage1_buildDependencyRequest struct {
		BaseRequest rootBuild_stage1_request
		DependencyPath string
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
				BaseRequest: req,
				DependencyPath: dependencyPath,
			})
		}
	
		return nil, buildDependencyRequests
	}
	
	var rootBuild_stage1_buildDependency = func (ctx *task.ExecutionContext, req rootBuild_stage1_buildDependencyRequest) (error, string) {
	
		// verify that root path is valid for dependency
		_, err := RootPath(req.DependencyPath, "")
		if err != nil {
			return err, ""
		}
	
		err, dependencyFingerprint := RootBuild(
			ctx,
			RootBuildRequest{
				Context: req.BaseRequest.Context,
				RootPath: req.DependencyPath,
			},
		)
		if err != nil {
			return err, ""
		}
	
		dependencyHeapPath := filepath.Join(req.BaseRequest.GardenPath, "dyd", "heap", "stems", dependencyFingerprint)
	
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
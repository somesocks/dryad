package core

import (
	// dydfs "dryad/filesystem"
	"dryad/task"

	// "io/fs"
	// "io/ioutil"
	// "os"
	// "path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage5_request struct {
	RelRootPath string
	RootStemPath string
	StemBuildPath string
	RootFingerprint string
}

// stage 5 - execute the root to build its stem,
var rootBuild_stage5 func (ctx *task.ExecutionContext, req rootBuild_stage5_request) (error, string) =
	func (ctx *task.ExecutionContext, req rootBuild_stage5_request) (error, string) {
		zlog.Debug().
			Str("path", req.RelRootPath).
			Msg("root build - stage5")

		var err error

		err = StemInit(req.StemBuildPath)
		if err != nil {
			return err, ""
		}
		err = StemRun(StemRunRequest{
			StemPath: req.RootStemPath,
			Env: map[string]string{
				"DYD_BUILD": req.StemBuildPath,
			},
			Args:       []string{req.StemBuildPath},
			JoinStdout: false,
		})
		if err != nil {
			return err, ""
		}

		// prepare the path
		err = rootBuild_pathPrepare(req.StemBuildPath)
		if err != nil {
			return err, ""
		}

		// prepare the requirements dir
		err = rootBuild_requirementsPrepare(req.StemBuildPath)
		if err != nil {
			return err, ""
		}

		err, stemBuildFingerprint := stemFinalize(ctx, req.StemBuildPath)
		if err != nil {
			return err, ""
		}

		return err, stemBuildFingerprint
	}

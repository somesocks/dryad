package core

import (
	// dydfs "dryad/filesystem"
	"dryad/task"

	// "io/fs"
	// "io/ioutil"
	// "os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage5_request struct {
	Garden *SafeGardenReference
	RelRootPath string
	RootStemPath string
	StemBuildPath string
	RootFingerprint string
	JoinStdout bool
	JoinStderr bool
}

// stage 5 - execute the root to build its stem,
var rootBuild_stage5 func (ctx *task.ExecutionContext, req rootBuild_stage5_request) (error, string) =
	func (ctx *task.ExecutionContext, req rootBuild_stage5_request) (error, string) {
		zlog.Trace().
			Str("path", req.RelRootPath).
			Msg("root build - stage5")

		var err error

		err = StemInit(req.StemBuildPath)
		if err != nil {
			return err, ""
		}
		err = StemRun(StemRunRequest{
			Garden: req.Garden,
			StemPath: req.RootStemPath,
			Env: map[string]string{
				"DYD_BUILD": req.StemBuildPath,
			},
			Args:       []string{req.StemBuildPath},
			JoinStdout: req.JoinStdout,
			JoinStderr: req.JoinStderr,
			MainOverride: filepath.Join(req.RootStemPath, "dyd", "commands", "dyd-root-build"),
		})
		if err != nil {
			zlog.Debug().
				Str("path", req.RelRootPath).
				Err(err).
				Msg("root build - stage5 - error executing root")
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

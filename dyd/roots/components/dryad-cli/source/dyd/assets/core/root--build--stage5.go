package core

import (
	// dydfs "dryad/filesystem"

	// "io/fs"
	// "io/ioutil"
	// "os"
	"context"
	"dryad/internal/filepath"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage5_request struct {
	Garden          *SafeGardenReference
	RelRootPath     string
	RootStemPath    string
	StemBuildPath   string
	RootFingerprint string
	JoinStdout      bool
	JoinStderr      bool
	LogStdout       struct {
		Path string
		Name string
	}
	LogStderr struct {
		Path string
		Name string
	}
}

// stage 5 - execute the root to build its stem,
var rootBuild_stage5 func(ctx context.Context, req rootBuild_stage5_request) (error, string) = func(ctx context.Context, req rootBuild_stage5_request) (error, string) {
	zlog.Trace().
		Str("path", req.RelRootPath).
		Msg("root build - stage5")

	var err error

	err = StemInit(req.StemBuildPath)
	if err != nil {
		return err, ""
	}
	err = StemRun(StemRunRequest{
		Garden:   req.Garden,
		StemPath: req.RootStemPath,
		Env: map[string]string{
			"DYD_BUILD": req.StemBuildPath,
		},
		Args:         []string{req.StemBuildPath},
		Stdin:        InputConfig{Mode: InputInherit},
		Stdout:       OutputConfigFromOptions(req.JoinStdout, req.LogStdout.Path, "dyd-root-build--"+sanitizePathSegment(req.RelRootPath)+".out"),
		Stderr:       OutputConfigFromOptions(req.JoinStderr, req.LogStderr.Path, "dyd-root-build--"+sanitizePathSegment(req.RelRootPath)+".err"),
		MainOverride: filepath.Join(req.RootStemPath, "dyd", "commands", "dyd-root-build"),
	})
	if err != nil {
		zlog.Debug().
			Str("path", req.RelRootPath).
			Err(err).
			Msg("root build - stage5 - error executing root")
		return err, ""
	}

	err = rootBuild_requirementsPrepare(req.StemBuildPath)
	if err != nil {
		return err, ""
	}

	// prepare the path
	err = rootBuild_pathPrepare(req.StemBuildPath)
	if err != nil {
		return err, ""
	}

	err, stemBuildFingerprint := stemFinalize(ctx, req.StemBuildPath)
	if err != nil {
		return err, ""
	}

	return err, stemBuildFingerprint
}

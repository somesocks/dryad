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

type rootBuild_stage0_request struct {
	RootPath string
	WorkspacePath string
}

// stage 0 - build a shallow partial clone of the root into a working directory,
// so we can build it into a stem
var rootBuild_stage0 = func () (func (ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, any)) {

	var prepReq = func (ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0")

		zlog.Trace().
			Msg("RootBuild/stage0/prepReq")

		return nil, req
	}

	var mkBaseDir = func (ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/mkBaseDir")

			err := os.MkdirAll(
			filepath.Join(req.WorkspacePath, "dyd"),
			os.ModePerm,
		)
		return err, req	
	}

	var linkAssetsDir = func (ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkAssetsDir")

		exists, err := fileExists(filepath.Join(req.RootPath, "dyd", "assets"))
		if err != nil {
			return err, req
		}
		if exists {
			err = os.Symlink(
				filepath.Join(req.RootPath, "dyd", "assets"),
				filepath.Join(req.WorkspacePath, "dyd", "assets"),
			)
			if err != nil {
				return err,  req
			}
		}

		return nil, req
	}

	var linkCommandsDir = func (ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkCommandsDir")

			exists, err := fileExists(filepath.Join(req.RootPath, "dyd", "commands"))
		if err != nil {
			return err, req
		}
		if exists {
			err = os.Symlink(
				filepath.Join(req.RootPath, "dyd", "commands"),
				filepath.Join(req.WorkspacePath, "dyd", "commands"),
			)
			if err != nil {
				return err, req
			}
		}

		return nil, req
	}

	var linkSecretsDir = func (ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkSecretsDir")

			exists, err := fileExists(filepath.Join(req.RootPath, "dyd", "secrets"))
		if err != nil {
			return err, req
		}
		if exists {
			err = os.Symlink(
				filepath.Join(req.RootPath, "dyd", "secrets"),
				filepath.Join(req.WorkspacePath, "dyd", "secrets"),
			)
			if err != nil {
				return err, req
			}
		}

		return nil, req
	}


	var linkTraitsDir = func (ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkTraitsDir")

			exists, err := fileExists(filepath.Join(req.RootPath, "dyd", "traits"))
		if err != nil {
			return err, req
		}
		if exists {
			err = os.Symlink(
				filepath.Join(req.RootPath, "dyd", "traits"),
				filepath.Join(req.WorkspacePath, "dyd", "traits"),
			)
			if err != nil {
				return err, req
			}
		}

		return nil, req
	}

	var linkDocsDir = func (ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkDocsDir")

			exists, err := fileExists(filepath.Join(req.RootPath, "dyd", "docs"))
		if err != nil {
			return err, req
		}
		if exists {
			err = os.Symlink(
				filepath.Join(req.RootPath, "dyd", "docs"),
				filepath.Join(req.WorkspacePath, "dyd", "docs"),
			)
			if err != nil {
				return err, req
			}
		}

		return nil, req
	}

	var mkDependenciesDir = func (ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/mkDependenciesDir")

			err := os.MkdirAll(
			filepath.Join(req.WorkspacePath, "dyd", "dependencies"),
			os.ModePerm,
		)
		return err, req	
	}

	var linkRequirementsDir = func (ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkRequirementsDir")

		requirementsPath := filepath.Join(req.RootPath, "dyd", "requirements")
		exists, err := fileExists(requirementsPath)
		if err != nil {
			return err, req
		}

		if exists {
			err = os.Symlink(
				requirementsPath,
				filepath.Join(req.WorkspacePath, "dyd", "requirements"),
			)
			if err != nil {
				return err, req
			}
		} else {
			err = os.MkdirAll(
				filepath.Join(req.WorkspacePath, "dyd", "requirements"),
				os.ModePerm,
			)
			if err != nil {
				return err, req
			}
		}

		return nil, req
	}

	var rootBuild_stage0 = task.Series4(
		prepReq,
		mkBaseDir,
		task.Parallel7(
			mkDependenciesDir,
			linkRequirementsDir,
			linkAssetsDir,
			linkCommandsDir,
			linkSecretsDir,
			linkTraitsDir,
			linkDocsDir,
		),
		func (
			ctx *task.ExecutionContext,
			req task.Tuple7[
				rootBuild_stage0_request,
				rootBuild_stage0_request,
				rootBuild_stage0_request,
				rootBuild_stage0_request,
				rootBuild_stage0_request,
				rootBuild_stage0_request,
				rootBuild_stage0_request,
			],
		) (error, any) {
			return nil, nil
		},
	)

	return rootBuild_stage0

}()

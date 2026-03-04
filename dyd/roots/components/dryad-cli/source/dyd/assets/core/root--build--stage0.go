package core

import (
	// dydfs "dryad/filesystem"
	"dryad/task"

	// "io/fs"
	// "io/ioutil"
	"dryad/internal/os"
	stdos "os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage0_request struct {
	RootPath                 string
	WorkspacePath            string
	VariantDescriptor        string
	SelectedAssetsPath       string
	SelectedCommandsPath     string
	SelectedSecretsPath      string
	SelectedDocsPath         string
	SelectedRequirementsPath string
}

// stage 0 - build a shallow partial clone of the root into a working directory,
// so we can build it into a stem
var rootBuild_stage0 = func() func(ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, any) {

	var prepReq = func(ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0")

		zlog.Trace().
			Msg("RootBuild/stage0/prepReq")

		err, selectedPaths := rootBuild_selectAssetsAndCommandsAndSecretsAndDocsAndRequirementsPaths(
			ctx,
			req.RootPath,
			req.VariantDescriptor,
		)
		if err != nil {
			return err, req
		}

		req.SelectedAssetsPath = selectedPaths.AssetsPath
		req.SelectedCommandsPath = selectedPaths.CommandsPath
		req.SelectedSecretsPath = selectedPaths.SecretsPath
		req.SelectedDocsPath = selectedPaths.DocsPath
		req.SelectedRequirementsPath = selectedPaths.RequirementsPath

		return nil, req
	}

	var mkBaseDir = func(ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/mkBaseDir")

		err := os.MkdirAll(
			filepath.Join(req.WorkspacePath, "dyd"),
			stdos.ModePerm,
		)
		return err, req
	}

	var linkAssetsDir = func(ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkAssetsDir")

		if req.SelectedAssetsPath == "" {
			return nil, req
		}

		err := os.Symlink(
			req.SelectedAssetsPath,
			filepath.Join(req.WorkspacePath, "dyd", "assets"),
		)
		if err != nil {
			return err, req
		}

		return nil, req
	}

	var linkCommandsDir = func(ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkCommandsDir")

		if req.SelectedCommandsPath == "" {
			return nil, req
		}

		err := os.Symlink(
			req.SelectedCommandsPath,
			filepath.Join(req.WorkspacePath, "dyd", "commands"),
		)
		if err != nil {
			return err, req
		}

		return nil, req
	}

	var linkSecretsDir = func(ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkSecretsDir")

		if req.SelectedSecretsPath == "" {
			return nil, req
		}

		err := os.Symlink(
			req.SelectedSecretsPath,
			filepath.Join(req.WorkspacePath, "dyd", "secrets"),
		)
		if err != nil {
			return err, req
		}

		return nil, req
	}

	var linkTraitsDir = func(ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkTraitsDir")

		err := rootBuild_materializeVariantTraits(
			ctx,
			req.RootPath,
			req.WorkspacePath,
			req.VariantDescriptor,
		)
		if err != nil {
			return err, req
		}

		return nil, req
	}

	var linkDocsDir = func(ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkDocsDir")

		if req.SelectedDocsPath == "" {
			return nil, req
		}

		err := os.Symlink(
			req.SelectedDocsPath,
			filepath.Join(req.WorkspacePath, "dyd", "docs"),
		)
		if err != nil {
			return err, req
		}

		return nil, req
	}

	var mkDependenciesDir = func(ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/mkDependenciesDir")

		err := os.MkdirAll(
			filepath.Join(req.WorkspacePath, "dyd", "dependencies"),
			stdos.ModePerm,
		)
		return err, req
	}

	var linkRequirementsDir = func(ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/linkRequirementsDir")

		if req.SelectedRequirementsPath == "" {
			return nil, req
		}

		err := os.Symlink(
			req.SelectedRequirementsPath,
			filepath.Join(req.WorkspacePath, "dyd", "~requirements"),
		)
		if err != nil {
			return err, req
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
		func(
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

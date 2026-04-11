package core

import (
	// dydfs "dryad/filesystem"
	"dryad/task"

	// "io/fs"
	// "io/ioutil"
	"dryad/internal/filepath"
	"dryad/internal/os"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage0_request struct {
	RootPath                 string
	WorkspacePath            string
	VariantDescriptor        string
	SelectedAssetsPath       string
	SelectedCommandsPath     string
	SelectedTraitsPath       string
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

		rootRef := SafeRootReference{BasePath: req.RootPath}
		err, unsafeVariant := rootRef.VariantFromFilesystem(req.VariantDescriptor)
		if err != nil {
			return err, req
		}

		err, variant := unsafeVariant.Resolve(ctx)
		if err != nil {
			return err, req
		}

		if variant.Assets != nil {
			req.SelectedAssetsPath = variant.Assets.BasePath
		}
		if variant.Commands != nil {
			req.SelectedCommandsPath = variant.Commands.BasePath
		}
		if variant.Traits != nil {
			req.SelectedTraitsPath = variant.Traits.BasePath
		}
		if variant.Secrets != nil {
			req.SelectedSecretsPath = variant.Secrets.BasePath
		}
		if variant.Docs != nil {
			req.SelectedDocsPath = variant.Docs.BasePath
		}
		if variant.Requirements != nil {
			req.SelectedRequirementsPath = variant.Requirements.BasePath
		}

		return nil, req
	}

	var mkBaseDir = func(ctx *task.ExecutionContext, req rootBuild_stage0_request) (error, rootBuild_stage0_request) {
		zlog.Trace().
			Msg("RootBuild/stage0/mkBaseDir")

		err := os.MkdirAll(
			filepath.Join(req.WorkspacePath, "dyd"),
			os.ModePerm,
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
			req.SelectedTraitsPath,
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
			os.ModePerm,
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

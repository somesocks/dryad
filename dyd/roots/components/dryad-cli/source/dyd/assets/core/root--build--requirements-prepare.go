package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"

	"os"
	"path/filepath"
	"regexp"

	zlog "github.com/rs/zerolog/log"
)

// This function is used to prepare the dyd/requirements for a package,
// based on the contents of dyd/dependencies.
var rootBuild_requirementsPrepare = func() func(string) error {

	var RE_ROOT_BUILD_REQUIREMENTS_PREPARE_SHOULD_CRAWL = regexp.MustCompile(
		"^(" +
			"([^/]*)" +
			"|([^/]*/dyd)" +
			"|([^/]*/dyd/traits)" +
			"|([^/]*/dyd/traits/.*)" +
			")$",
	)

	// crawler used to match files to copy
	// node.BasePath should be the path to the package dyd/dependencies
	var fs_should_walk = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		var relPath, relErr = filepath.Rel(node.BasePath, node.VPath)
		if relErr != nil {
			return relErr, false
		}
		shouldCrawl := RE_ROOT_BUILD_REQUIREMENTS_PREPARE_SHOULD_CRAWL.MatchString(relPath)
		zlog.Trace().
			Str("node.VPath", node.VPath).
			Bool("shouldCrawl", shouldCrawl).
			Msg("rootBuild_requirementsPrepare.fs_should_walk")
		return nil, shouldCrawl
	}

	var RE_ROOT_BUILD_REQUIREMENTS_PREPARE_SHOULD_MATCH = regexp.MustCompile(
		"^(" +
			"([^/]*/dyd/fingerprint)" +
			"|([^/]*/dyd/traits/.*)" +
			")$",
	)

	// matcher used to match files to copy
	// node.BasePath should be the path to the package dyd/dependencies
	var fs_should_match = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		var relPath, relErr = filepath.Rel(node.BasePath, node.VPath)
		if relErr != nil {
			return relErr, false
		}
		shouldMatch := RE_ROOT_BUILD_REQUIREMENTS_PREPARE_SHOULD_MATCH.MatchString(relPath)
		zlog.Trace().
			Str("node.VPath", node.VPath).
			Bool("shouldMatch", shouldMatch).
			Msg("rootBuild_requirementsPrepare.fs_should_match")
		return nil, shouldMatch
	}

	var fs_on_match = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		zlog.Trace().
			Str("node.VPath", node.VPath).
			Msg("rootBuild_requirementsPrepare.fs_on_match")

		relPath, err := filepath.Rel(node.BasePath, node.VPath)
		if err != nil {
			return err, nil
		}
		// zlog.Trace().
		// 	Str("relPath", relPath).
		// 	Msg("rootBuild_requirementsPrepare.fs_on_match")

		rootPath := filepath.Dir(filepath.Dir(node.BasePath))

		reqsPath := filepath.Join(rootPath, "dyd", "requirements", relPath)
		// zlog.Trace().
		// 	Str("reqsPath", reqsPath).
		// 	Msg("rootBuild_requirementsPrepare.fs_on_match")

		reqsParentPath := filepath.Dir(reqsPath)
		err = os.MkdirAll(reqsParentPath, os.ModePerm)
		if err != nil {
			return err, nil
		}

		// zlog.Trace().
		// 	Str("reqsParentPath", reqsParentPath).
		// 	Msg("rootBuild_requirementsPrepare.fs_on_match")

		var isSymlink = node.Info.Mode()&os.ModeSymlink == os.ModeSymlink
		var isInternalLink = false
		var linkTarget = ""

		// zlog.Trace().
		// 	Bool("isSymlink", isSymlink).
		// 	Msg("rootBuild_requirementsPrepare.fs_on_match")

		// check if its an package-internal symlink
		if isSymlink {
			linkTarget, err = os.Readlink(node.Path)
			if err != nil {
				return err, nil
			}

			absLinkTarget := linkTarget

			// clean up relative links
			if !filepath.IsAbs(absLinkTarget) {
				absLinkTarget = filepath.Clean(filepath.Join(filepath.Dir(node.Path), absLinkTarget))
			}

			isInternalLink, err = fileIsDescendant(absLinkTarget, node.BasePath)
			if err != nil {
				return err, nil
			}
		}

		// if it's an internal link clone it,
		// otherwise, copy the file
		if isInternalLink {
			err = os.Symlink(linkTarget, reqsPath)
			if err != nil {
				return err, nil
			}
		} else {
			srcFile, err := os.Open(node.Path)
			if err != nil {
				return err, nil
			}
			defer srcFile.Close()

			var destFile *os.File
			destFile, err = os.Create(reqsPath)
			if err != nil {
				return err, nil
			}
			defer destFile.Close()

			_, err = destFile.ReadFrom(srcFile)
			if err != nil {
				return err, nil
			}

			// heap files should be set to R-X--X--X
			err = destFile.Chmod(0o511)
			if err != nil {
				return err, nil
			}

		}

		return nil, nil
	}

	fs_on_match = dydfs.ConditionalWalkAction(fs_on_match, fs_should_match)

	var action = func(workspacePath string) error {
		zlog.Trace().
			Str("workspacePath", workspacePath).
			Msg("rootBuild_requirementsPrepare")

		requirementsPath := filepath.Join(workspacePath, "dyd", "requirements")

		err, _ := dydfs.RemoveAll(task.SERIAL_CONTEXT, requirementsPath)
		if err != nil {
			zlog.Trace().
				Err(err).
				Msg("rootBuild_requirementsPrepare RemoveAll err")
			return err
		}

		dependenciesPath := filepath.Join(workspacePath, "dyd", "dependencies")

		// NOTE: this should use a serial execution context until it has been verified to be
		// concurrent-safe
		err, _ = dydfs.Walk6(
			task.SERIAL_CONTEXT,
			dydfs.Walk6Request{
				BasePath:   dependenciesPath,
				Path:       dependenciesPath,
				VPath:      dependenciesPath,
				ShouldWalk: fs_should_walk,
				OnPreMatch: fs_on_match,
			})
		if err != nil {
			return err
		}

		return nil
	}

	return action

}()

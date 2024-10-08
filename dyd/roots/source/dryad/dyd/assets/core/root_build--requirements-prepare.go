package core

import (
	dydfs "dryad/filesystem"

	"os"
	"path/filepath"
	"regexp"

	zlog "github.com/rs/zerolog/log"
)

var RE_ROOT_BUILD_REQUIREMENTS_PREPARE_SHOULD_CRAWL = regexp.MustCompile(
	"^(" +
		"([^/]*)" +
		"|([^/]*/dyd)" +
		"|([^/]*/dyd/traits)" +
		"|([^/]*/dyd/traits/.*)" +
		")$",
)

var RE_ROOT_BUILD_REQUIREMENTS_PREPARE_SHOULD_MATCH = regexp.MustCompile(
	"^(" +
		"([^/]*/dyd/fingerprint)" +
		"|([^/]*/dyd/secrets-fingerprint)" +
		"|([^/]*/dyd/traits/.*)" +
		")$",
)


// This function is used to prepare the dyd/requirements for a package,
// based on the contents of dyd/dependencies.
var rootBuild_requirementsPrepare = func() func(string) error {

	// crawler used to match files to copy 
	// context.BasePath should be the path to the package dyd/dependencies 
	var fs_should_crawl = func(context dydfs.Walk4Context) (bool, error) {
		var relPath, relErr = filepath.Rel(context.BasePath, context.VPath)
		if relErr != nil {
			return false, relErr
		}
		shouldCrawl := RE_ROOT_BUILD_REQUIREMENTS_PREPARE_SHOULD_CRAWL.MatchString(relPath)
		zlog.Trace().
			Str("context.VPath", context.VPath).
			Bool("shouldCrawl", shouldCrawl).
			Msg("rootBuild_requirementsPrepare.fs_should_crawl")
		return shouldCrawl, nil
	}

	// matcher used to match files to copy 
	// context.BasePath should be the path to the package dyd/dependencies 
	var fs_should_match = func(context dydfs.Walk4Context) (bool, error) {
		var relPath, relErr = filepath.Rel(context.BasePath, context.VPath)
		if relErr != nil {
			return false, relErr
		}
		shouldMatch := RE_ROOT_BUILD_REQUIREMENTS_PREPARE_SHOULD_MATCH.MatchString(relPath)
		zlog.Trace().
			Str("context.VPath", context.VPath).
			Bool("shouldMatch", shouldMatch).
			Msg("rootBuild_requirementsPrepare.fs_should_match")
		return shouldMatch, nil
	}

	var fs_on_match = func(context dydfs.Walk4Context) error {
		zlog.Trace().
			Str("context.VPath", context.VPath).
			Msg("rootBuild_requirementsPrepare.fs_on_match")

		relPath, err := filepath.Rel(context.BasePath, context.VPath)
		if err != nil {
			return err
		}
		// zlog.Trace().
		// 	Str("relPath", relPath).
		// 	Msg("rootBuild_requirementsPrepare.fs_on_match")

		rootPath := filepath.Dir(filepath.Dir(context.BasePath))

		reqsPath := filepath.Join(rootPath, "dyd", "requirements", relPath)
		// zlog.Trace().
		// 	Str("reqsPath", reqsPath).
		// 	Msg("rootBuild_requirementsPrepare.fs_on_match")

		reqsParentPath := filepath.Dir(reqsPath)
		err = os.MkdirAll(reqsParentPath, os.ModePerm)
		if err != nil {
			return err
		}

		// zlog.Trace().
		// 	Str("reqsParentPath", reqsParentPath).
		// 	Msg("rootBuild_requirementsPrepare.fs_on_match")

		var isSymlink = context.Info.Mode()&os.ModeSymlink == os.ModeSymlink
		var isInternalLink = false
		var linkTarget = ""

		// zlog.Trace().
		// 	Bool("isSymlink", isSymlink).
		// 	Msg("rootBuild_requirementsPrepare.fs_on_match")

		// check if its an package-internal symlink
		if isSymlink {
			linkTarget, err = os.Readlink(context.Path)
			if err != nil {
				return err
			}

			absLinkTarget := linkTarget

			// clean up relative links
			if !filepath.IsAbs(absLinkTarget) {
				absLinkTarget = filepath.Clean(filepath.Join(filepath.Dir(context.Path), absLinkTarget))
			}

			isInternalLink, err = fileIsDescendant(absLinkTarget, context.BasePath)
			if err != nil {
				return err
			}
		}

		// if it's an internal link clone it,
		// otherwise, copy the file
		if isInternalLink {
			err = os.Symlink(linkTarget, reqsPath)
			if err != nil {
				return err
			}
		} else {
			srcFile, err := os.Open(context.Path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			var destFile *os.File
			destFile, err = os.Create(reqsPath)
			if err != nil {
				return err
			}
			defer destFile.Close()

			_, err = destFile.ReadFrom(srcFile)
			if err != nil {
				return err
			}

			// heap files should be set to R-X--X--X
			err = destFile.Chmod(0o511)
			if err != nil {
				return err
			}

		}

		return nil
	}

	var action = func(workspacePath string) error {
		zlog.Trace().
			Str("workspacePath", workspacePath).
			Msg("rootBuild_requirementsPrepare")

		requirementsPath := filepath.Join(workspacePath, "dyd", "requirements")

		err := dydfs.RemoveAll(requirementsPath)
		if err != nil {
			return err
		}

		dependenciesPath := filepath.Join(workspacePath, "dyd", "dependencies")

		err = dydfs.BFSWalk2(dydfs.Walk4Request{
			BasePath:    dependenciesPath,
			Path:        dependenciesPath,
			VPath:       dependenciesPath,
			ShouldCrawl: fs_should_crawl,
			ShouldMatch: fs_should_match,
			OnMatch:     fs_on_match,
		})
		if err != nil {
			return err
		}

		return nil
	}

	return action

}()

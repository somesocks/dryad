package core

import (
	fs2 "dryad/filesystem"
	"os"
	"path/filepath"
	"regexp"

	zlog "github.com/rs/zerolog/log"
)

var rootBuild_requirementsPrepare = func() func(string) error {

	var re_should_crawl = regexp.MustCompile(
		"^((.*/dyd)|(.*/dyd/dependencies)|(.*/dyd/dependencies/[^/]*)|(.*/dyd/dependencies/[^/]*/dyd)|(.*/dyd/dependencies/.*/dyd/traits)|(.*/dyd/dependencies/[^/]*/dyd/traits/.*))$",
	)

	var fs_should_crawl = func(context fs2.Walk4Context) (bool, error) {
		shouldCrawl := re_should_crawl.MatchString(context.VPath)
		zlog.Trace().
			Str("context.VPath", context.VPath).
			Bool("shouldCrawl", shouldCrawl).
			Msg("rootBuild_requirementsPrepare.fs_should_crawl")
		return shouldCrawl, nil
	}

	// var fs_on_crawl = func(context fs2.Walk4Context) error {
	// 	fmt.Println("fs_on_crawl", context.VPath)
	// 	return nil
	// }

	var re_should_match = regexp.MustCompile("^(.*/dyd/dependencies/[^/]*/dyd/fingerprint)|(.*/dyd/dependencies/[^/]*/dyd/secrets-fingerprint)|(.*/dyd/dependencies/[^/]*/dyd/traits/.*)$")

	var fs_should_match = func(context fs2.Walk4Context) (bool, error) {
		shouldMatch := re_should_match.MatchString(context.VPath)
		zlog.Trace().
			Str("context.VPath", context.VPath).
			Bool("shouldMatch", shouldMatch).
			Msg("rootBuild_requirementsPrepare.fs_should_match")
		return shouldMatch, nil
	}

	var fs_on_match = func(context fs2.Walk4Context) error {
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

			err = destFile.Sync()
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

		err := os.RemoveAll(requirementsPath)
		if err != nil {
			return err
		}

		dependenciesPath := filepath.Join(workspacePath, "dyd", "dependencies")

		err = fs2.BFSWalk2(fs2.Walk4Request{
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

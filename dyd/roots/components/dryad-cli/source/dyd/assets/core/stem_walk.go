package core

import (
	fs2 "dryad/filesystem"
	"os"
	"path/filepath"
	"regexp"

	zlog "github.com/rs/zerolog/log"
)

var RE_STEM_WALK_SHOULD_CRAWL = regexp.MustCompile(
	"^(" +
		"(\\.)" +
		"|(dyd)" +
		"|(dyd/path)" +
		"|(dyd/assets)" +
		"|(dyd/assets/.*)" +
		"|(dyd/commands)" +
		"|(dyd/commands/.*)" +
		"|(dyd/docs)" +
		"|(dyd/docs/.*)" +
		"|(dyd/requirements)" +
		"|(dyd/requirements/.*)" +
		"|(dyd/traits)" +
		"|(dyd/traits/.*)" +
		"|(dyd/dependencies)" +
		")$",
)

// should walk

// - if the vpath does not match the pattern then no
// - else if the node is a symlink then
// 	- if the node is a symlink pointing to a relative location within the package then no
// 	- else yes
// - else if the node is a directory then yes
// - else if the node is a file then no
// - else error?
func StemWalkShouldCrawl(context fs2.Walk4Context) (bool, error) {
	zlog.
		Trace().
		Str("path", context.Path).
		Str("vPath", context.VPath).
		Str("basePath", context.BasePath).
		Msg("StemWalk / shouldCrawl")

	var relPath, relErr = filepath.Rel(context.BasePath, context.VPath)
	if relErr != nil {
		return false, relErr
	}
	matchesPath := RE_STEM_WALK_SHOULD_CRAWL.Match([]byte(relPath))

	if !matchesPath {
		return false, nil
	} else if context.Info.IsDir() {
		return true, nil
	} else if context.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
		linkTarget, err := os.Readlink(context.Path)
		if err != nil {
			return false, err
		}

		// clean up relative links
		absLinkTarget := linkTarget
		if !filepath.IsAbs(absLinkTarget) {
			absLinkTarget = filepath.Join(filepath.Dir(context.VPath), linkTarget)
		} 

		isDescendant, err := fileIsDescendant(absLinkTarget, context.BasePath)

		if err != nil {
			return false, err
		}

		return !isDescendant, nil
	} else {
		return false, nil
	}
}

var RE_STEM_WALK_SHOULD_MATCH = regexp.MustCompile(
	"^(" +
		"(dyd)" +
		"|(dyd/path)" +
		"|(dyd/path/.*)" +
		"|(dyd/assets)" +
		"|(dyd/assets/.*)" +
		"|(dyd/commands)" +
		"|(dyd/commands/.*)" +
		"|(dyd/docs)" +
		"|(dyd/docs/.*)" +
		"|(dyd/type)" +
		"|(dyd/fingerprint)" +
		"|(dyd/secrets-fingerprint)" +
		"|(dyd/requirements)" +
		"|(dyd/requirements/.*)" +
		"|(dyd/dependencies)" +
		"|(dyd/traits)" +
		"|(dyd/traits/.*)" +
		")$",
)

// should match
// - if the vpath does not match the pattern then no,
// - else if the node is a symlink then
// 	- if the node is a symlink pointing to a relative location within the package then yes,
// 	- else no
// - else if the node is a directory then no,
// - else if the node is a file then yes,
// - else error?
func StemWalkShouldMatch(context fs2.Walk4Context) (bool, error) {
	zlog.
		Trace().
		Str("path", context.Path).
		Str("vPath", context.VPath).
		Str("basePath", context.BasePath).
		Msg("StemWalk / shouldMatch")

	var relPath, relErr = filepath.Rel(context.BasePath, context.VPath)
	if relErr != nil {
		return false, relErr
	}
	matchesPath := RE_STEM_WALK_SHOULD_MATCH.Match([]byte(relPath))
	return matchesPath, nil
}

type StemWalkRequest struct {
	BasePath string
	OnMatch  func(context fs2.Walk4Context) error
}

func StemWalk(args StemWalkRequest) error {
	var path string
	var err error

	path, err = filepath.EvalSymlinks(args.BasePath)
	if err != nil {
		return err
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}

	return fs2.BFSWalk2(fs2.Walk4Request{
		Path:        path,
		VPath:       path,
		BasePath:    path,
		ShouldCrawl: StemWalkShouldCrawl,
		ShouldMatch: StemWalkShouldMatch,
		OnMatch:     args.OnMatch,
	})
}

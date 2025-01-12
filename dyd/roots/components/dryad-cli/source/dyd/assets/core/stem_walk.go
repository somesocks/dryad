package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

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
func StemWalkShouldCrawl(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
	zlog.
		Trace().
		Str("path", node.Path).
		Str("vPath", node.VPath).
		Str("basePath", node.BasePath).
		Msg("StemWalk / shouldCrawl")

	var relPath, relErr = filepath.Rel(node.BasePath, node.VPath)
	if relErr != nil {
		return relErr, false
	}
	matchesPath := RE_STEM_WALK_SHOULD_CRAWL.Match([]byte(relPath))

	if !matchesPath {
		return nil, false 
	} else if node.Info.IsDir() {
		return nil, true
	} else if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
		linkTarget, err := os.Readlink(node.Path)
		if err != nil {
			return err, false
		}

		// clean up relative links
		absLinkTarget := linkTarget
		if !filepath.IsAbs(absLinkTarget) {
			absLinkTarget = filepath.Join(filepath.Dir(node.VPath), linkTarget)
		} 

		isDescendant, err := fileIsDescendant(absLinkTarget, node.BasePath)

		if err != nil {
			return err, false
		}

		return  nil, !isDescendant
	} else {
		return nil, false 
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
func StemWalkShouldMatch(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
	zlog.
		Trace().
		Str("path", node.Path).
		Str("vPath", node.VPath).
		Str("basePath", node.BasePath).
		Msg("StemWalk / shouldMatch")

	var relPath, relErr = filepath.Rel(node.BasePath, node.VPath)
	if relErr != nil {
		return relErr, false
	}
	matchesPath := RE_STEM_WALK_SHOULD_MATCH.Match([]byte(relPath))
	return nil, matchesPath 
}

type StemWalkRequest struct {
	BasePath string
	OnMatch  func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any)
}

func StemWalk(
	ctx *task.ExecutionContext,
	args StemWalkRequest,
) error {
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

	err, _ = fs2.BFSWalk3(
		ctx,
		fs2.Walk5Request{
			Path:        path,
			VPath:       path,
			BasePath:    path,
			ShouldCrawl: StemWalkShouldCrawl,
			ShouldMatch: StemWalkShouldMatch,
			OnMatch:     args.OnMatch,
			OnError: func(err error, node fs2.Walk5Node) error {
				zlog.
					Trace().
					Str("path", node.Path).
					Str("vpath", node.VPath).
					Err(err).
					Msg("StemWalk / onError")
				
				return err
			},
		},
	)

	zlog.
		Trace().
		Err(err).
		Msg("StemWalk / err")

	return err
}

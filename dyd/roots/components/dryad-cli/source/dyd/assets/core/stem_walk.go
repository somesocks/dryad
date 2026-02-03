package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"

	"os"
	"path/filepath"
	"regexp"

	zlog "github.com/rs/zerolog/log"
)

type DydIgnoreRequest struct {
	BasePath string
	Path     string
}

var readDydIgnore task.Task[DydIgnoreRequest, *dydfs.GlobMatcher] = func() task.Task[DydIgnoreRequest, *dydfs.GlobMatcher] {
	var read task.Task[DydIgnoreRequest, *dydfs.GlobMatcher]

	read = func(ctx *task.ExecutionContext, req DydIgnoreRequest) (error, *dydfs.GlobMatcher) {
		var parentMatcher *dydfs.GlobMatcher
		var err error

		if req.Path != req.BasePath {
			parentDir := filepath.Dir(req.Path)
			err, parentMatcher = read(ctx, DydIgnoreRequest{
				BasePath: req.BasePath,
				Path:     parentDir,
			})
			if err != nil {
				return err, nil
			}
		}

		var ignoreFile string
		var matcher *dydfs.GlobMatcher

		ignoreFile = filepath.Join(req.Path, ".dyd-ignore")

		zlog.
			Trace().
			Str("ignoreFile", ignoreFile).
			Msg("StemWalk / readDydIgnore")

		err, matcher = dydfs.NewGlobMatcherFromFile(ignoreFile, parentMatcher)
		return err, matcher
	}

	read = task.Memoize(
		read,
		func(ctx *task.ExecutionContext, req DydIgnoreRequest) (error, any) {
			return nil, "readDydIgnore" + "-" + req.BasePath + "-" + req.Path
		},
	)

	return read
}()

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
		"|(dyd/secrets)" +
		"|(dyd/secrets/.*)" +
		"|(dyd/requirements)" +
		"|(dyd/requirements/.*)" +
		"|(dyd/traits)" +
		"|(dyd/traits/.*)" +
		"|(dyd/dependencies)" +
		")$",
)

var RE_STEM_WALK_SHOULD_CHECK_DYD_IGNORE = regexp.MustCompile(
	"^(" +
		"|(dyd/assets)" +
		"|(dyd/assets/.*)" +
		")$",
)

// should walk

// - if the vpath does not match the pattern then no
// - else if the node is a symlink then
//   - if the node is a symlink pointing to a relative location within the package then no
//   - else yes
//
// - else if the node is a directory then yes
// - else if the node is a file then no
// - else error?
func StemWalkShouldCrawl(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
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
	shouldCheckDydIgnore := RE_STEM_WALK_SHOULD_CHECK_DYD_IGNORE.Match([]byte(relPath))

	zlog.
		Trace().
		Str("path", node.Path).
		Str("vPath", node.VPath).
		Str("basePath", node.BasePath).
		Bool("matchesPath", matchesPath).
		Bool("shouldCheckDydIgnore", shouldCheckDydIgnore).
		Msg("StemWalk / shouldCrawl 2")

	if !matchesPath {
		return nil, false
	} else if node.Info.IsDir() {
		if shouldCheckDydIgnore {
			parentDir := filepath.Dir(node.VPath)

			err, matcher := readDydIgnore(ctx, DydIgnoreRequest{
				BasePath: node.BasePath,
				Path:     parentDir,
			})
			if err != nil {
				return err, false
			}

			err, match := matcher.Match(dydfs.NewGlobPath(node.VPath, true))

			zlog.
				Trace().
				Str("path", node.Path).
				Str("vPath", node.VPath).
				Str("basePath", node.BasePath).
				Bool("match", match).
				Msg("StemWalk / shouldCrawl dydIgnore match")

			if err != nil {
				return err, false
			} else if match {
				return nil, false
			}
		}

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

		return nil, !isDescendant
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
		"|(dyd/secrets)" +
		"|(dyd/secrets/.*)" +
		"|(dyd/type)" +
		"|(dyd/fingerprint)" +
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
//   - if the node is a symlink pointing to a relative location within the package then yes,
//   - else no
//
// - else if the node is a directory then no,
// - else if the node is a file then yes,
// - else error?
func StemWalkShouldMatch(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
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
	shouldCheckDydIgnore := RE_STEM_WALK_SHOULD_CHECK_DYD_IGNORE.Match([]byte(relPath))

	if matchesPath && shouldCheckDydIgnore {
		parentDir := filepath.Dir(node.VPath)

		err, matcher := readDydIgnore(ctx, DydIgnoreRequest{
			BasePath: node.BasePath,
			Path:     parentDir,
		})
		if err != nil {
			return err, false
		}

		err, match := matcher.Match(dydfs.NewGlobPath(node.VPath, node.Info.IsDir()))

		zlog.
			Trace().
			Str("path", node.Path).
			Str("vPath", node.VPath).
			Str("basePath", node.BasePath).
			Bool("match", match).
			Msg("StemWalk / shouldMatch dydIgnore match")

		if err != nil {
			return err, false
		} else if match {
			return nil, false
		}
	}

	return nil, matchesPath
}

type StemWalkRequest struct {
	BasePath string
	OnMatch  func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any)
}

var StemWalk task.Task[StemWalkRequest, any] = func() task.Task[StemWalkRequest, any] {
	var stemWalk = func(
		ctx *task.ExecutionContext,
		args StemWalkRequest,
	) (error, any) {
		var path string
		var err error

		path, err = filepath.EvalSymlinks(args.BasePath)
		if err != nil {
			return err, nil
		}

		path, err = filepath.Abs(path)
		if err != nil {
			return err, nil
		}

		onMatch := dydfs.ConditionalWalkAction(
			args.OnMatch,
			StemWalkShouldMatch,
		)

		err, _ = dydfs.Walk6(
			ctx,
			dydfs.Walk6Request{
				BasePath:   path,
				Path:       path,
				VPath:      path,
				ShouldWalk: StemWalkShouldCrawl,
				OnPreMatch: onMatch,
			},
		)

		zlog.
			Trace().
			Err(err).
			Msg("StemWalk / err")

		return err, nil
	}

	// we want to replace the execution context, but with the same concurrency channel as before.
	// only the execution cache is replaced, to limit the scope of memoized calls to fetch dyd-ignore files
	stemWalk = task.WithContext(
		stemWalk,
		func(ctx *task.ExecutionContext, args StemWalkRequest) (error, *task.ExecutionContext) {
			return nil, &task.ExecutionContext{
				ConcurrencyChannel: ctx.ConcurrencyChannel,
			}
		},
	)

	return stemWalk
}()

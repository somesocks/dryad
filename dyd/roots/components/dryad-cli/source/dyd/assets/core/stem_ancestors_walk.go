package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	// "os"
	"path/filepath"
	"regexp"
	zlog "github.com/rs/zerolog/log"
)

var RE_STEM_ANCESTORS_SHOULD_CRAWL = regexp.MustCompile(
	"^(" +
		"(\\.)" +
		"|(dyd)" +
		"|(dyd/dependencies)" +
		"|((dyd/dependencies/[^/]*))" +
		"|((dyd/dependencies/[^/]*/)+dyd)" +
		"|((dyd/dependencies/[^/]*/)+dyd/dependencies)" +
		"|((dyd/dependencies/[^/]*/)+dyd/dependencies/[^/]*)" +
		")$",
)

// should walk

func stemAncestorsShouldWalk() func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, bool) {

	var crawlMap = make(map[string]bool)

	// - if the vpath matches the pattern then yes
	return func (ctx *task.ExecutionContext, node fs2.Walk6Node) (error, bool) {
		// don't crawl a path that's already been crawled
		if _, seen := crawlMap[node.Path]; seen {
			return nil, false
		}

		var relPath, relErr = filepath.Rel(node.BasePath, node.VPath)
		if relErr != nil {
			return relErr, false
		}
		matchesPath := RE_STEM_ANCESTORS_SHOULD_CRAWL.Match([]byte(relPath))

		isDir := node.Info.IsDir()
		if isDir {
			crawlMap[node.Path] = true
		}

		zlog.Trace().
			Str("relPath", relPath).
			Bool("crawl", matchesPath).
			Msg("stem ancestors walk / should crawl")

		return nil, matchesPath
	}
}

var stemAncestorsShouldMatch fs2.WalkDecision = func () fs2.WalkDecision {
		RE_STEM_ANCESTORS_WALK_SHOULD_MATCH := regexp.MustCompile(
			"^(" +
				"(\\.)" +
				"|((dyd/dependencies/[^/]*))" +
				"|((dyd/dependencies/[^/]*/)+dyd/dependencies/[^/]*)" +
			")$",
		)

		shouldMatch :=  func (ctx *task.ExecutionContext, node fs2.Walk6Node) (error, bool) {

			var relPath, relErr = filepath.Rel(node.BasePath, node.VPath)
			if relErr != nil {
				return relErr, false
			}
			matchesPath := RE_STEM_ANCESTORS_WALK_SHOULD_MATCH.Match([]byte(relPath))

			isDir := node.Info.IsDir()

			zlog.Trace().
				Str("relPath", relPath).
				Bool("match", matchesPath).
				Bool("isDir", isDir).
				Msg("stem ancestors walk / should match")

			return nil, matchesPath && isDir

		}

		return shouldMatch
	}()

type StemAncestorsWalkRequest struct {
	BasePath string
	OnMatch  func(node fs2.Walk6Node) error
	Self bool
}

func StemAncestorsWalk(args StemAncestorsWalkRequest) error {

	self := args.Self
	seen := 0
	onMatch := func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, any) {
		// skip the first match if we don't want to see ourselves
		seen += 1
		if !self && (seen == 1) {
			return nil, nil
		}

		return args.OnMatch(node), nil
	}

	onMatch = fs2.ConditionalWalkAction(
		onMatch,
		stemAncestorsShouldMatch,
	)
	
	// NOTE: this should run serially until checked for concurrency issues
	err, _ := fs2.Walk6(
		task.SERIAL_CONTEXT,
		fs2.Walk6Request{
			BasePath:    args.BasePath,
			Path:        args.BasePath,
			VPath:       args.BasePath,
			ShouldWalk: stemAncestorsShouldWalk(),
			OnPreMatch:     onMatch,
		},
	)

	return err
}

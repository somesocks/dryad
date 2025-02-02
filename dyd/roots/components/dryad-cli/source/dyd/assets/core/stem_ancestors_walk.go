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

func StemAncestorsWalkCrawler() func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {

	var crawlMap = make(map[string]bool)

	// - if the vpath matches the pattern then yes
	return func (ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
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



var RE_STEM_ANCESTORS_WALK_SHOULD_MATCH = regexp.MustCompile(
	"^(" +
		"(\\.)" +
		"|((dyd/dependencies/[^/]*))" +
		"|((dyd/dependencies/[^/]*/)+dyd/dependencies/[^/]*)" +
	")$",
)

// should match
func StemAncestorsWalkShouldMatch(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {

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

type StemAncestorsWalkRequest struct {
	BasePath string
	OnMatch  func(node fs2.Walk5Node) error
	Self bool
}

func StemAncestorsWalk(args StemAncestorsWalkRequest) error {

	self := args.Self
	seen := 0
	onMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
		// skip the first match if we don't want to see ourselves
		seen += 1
		if !self && (seen == 1) {
			return nil, nil
		}

		return args.OnMatch(node), nil
	}

	// NOTE: this should run serially until checked for concurrency issues
	err, _ := fs2.BFSWalk3(
		task.SERIAL_CONTEXT,
		fs2.Walk5Request{
			Path:        args.BasePath,
			VPath:       args.BasePath,
			BasePath:    args.BasePath,
			ShouldCrawl: StemAncestorsWalkCrawler(),
			ShouldMatch: StemAncestorsWalkShouldMatch,
			OnMatch:     onMatch,
		},
	)

	return err
}

package core

import (
	fs2 "dryad/filesystem"
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

func StemAncestorsWalkCrawler() func(fs2.Walk4Context) (bool, error) {

	var crawlMap = make(map[string]bool)

	// - if the vpath matches the pattern then yes
	return func (context fs2.Walk4Context) (bool, error) {
		// don't crawl a path that's already been crawled
		if _, seen := crawlMap[context.Path]; seen {
			return false, nil
		}

		var relPath, relErr = filepath.Rel(context.BasePath, context.VPath)
		if relErr != nil {
			return false, relErr
		}
		matchesPath := RE_STEM_ANCESTORS_SHOULD_CRAWL.Match([]byte(relPath))

		isDir := context.Info.IsDir()
		if isDir {
			crawlMap[context.Path] = true
		}

		zlog.Trace().
			Str("relPath", relPath).
			Bool("crawl", matchesPath).
			Msg("stem ancestors walk / should crawl")

		return matchesPath, nil
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
func StemAncestorsWalkShouldMatch(context fs2.Walk4Context) (bool, error) {

	var relPath, relErr = filepath.Rel(context.BasePath, context.VPath)
	if relErr != nil {
		return false, relErr
	}
	matchesPath := RE_STEM_ANCESTORS_WALK_SHOULD_MATCH.Match([]byte(relPath))

	isDir := context.Info.IsDir()


	zlog.Trace().
		Str("relPath", relPath).
		Bool("match", matchesPath).
		Bool("isDir", isDir).
		Msg("stem ancestors walk / should match")

	return matchesPath && isDir, nil

}

type StemAncestorsWalkRequest struct {
	BasePath string
	OnMatch  func(context fs2.Walk4Context) error
	Self bool
}

func StemAncestorsWalk(args StemAncestorsWalkRequest) error {

	self := args.Self
	seen := 0
	onMatch := func(context fs2.Walk4Context) error {
		// skip the first match if we don't want to see ourselves
		seen += 1
		if !self && (seen == 1) {
			return nil
		}

		return args.OnMatch(context)
	}

	return fs2.BFSWalk2(fs2.Walk4Request{
		Path:        args.BasePath,
		VPath:       args.BasePath,
		BasePath:    args.BasePath,
		ShouldCrawl: StemAncestorsWalkCrawler(),
		ShouldMatch: StemAncestorsWalkShouldMatch,
		OnMatch:     onMatch,
	})
}

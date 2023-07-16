package core

import (
	fs2 "dryad/filesystem"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

var RE_STEM_WALK_SHOULD_CRAWL = regexp.MustCompile(
	"^(" +
		"(\\.)" +
		"|(dyd)" +
		"|(dyd/path)" +
		"|(dyd/assets)" +
		"|(dyd/assets/.*)" +
		"|(dyd/traits)" +
		"|(dyd/traits/.*)" +
		"|(dyd/stems)" +
		"|(dyd/stems/[^/]*)" +
		"|(dyd/stems/.*/dyd)" +
		"|(dyd/stems/.*/dyd/traits(/.*)?)" +
		")$",
)

func StemWalkShouldCrawl(path string, info fs.FileInfo, basePath string) (bool, error) {
	var relPath, relErr = filepath.Rel(basePath, path)
	if relErr != nil {
		return false, relErr
	}
	matchesPath := RE_STEM_WALK_SHOULD_CRAWL.Match([]byte(relPath))
	isSymlink := info.Mode()&os.ModeSymlink == os.ModeSymlink
	shouldCrawl := matchesPath && !isSymlink
	return shouldCrawl, nil
}

var RE_STEM_WALK_SHOULD_MATCH = regexp.MustCompile(
	"^(" +
		"(dyd/path/.*)" +
		"|(dyd/assets/.*)" +
		"|(dyd/readme)" +
		"|(dyd/type)" +
		"|(dyd/fingerprint)" +
		"|(dyd/secrets-fingerprint)" +
		"|(dyd/main)" +
		"|(dyd/stems/.*/dyd/fingerprint)" +
		"|(dyd/stems/.*/dyd/traits/.*)" +
		"|(dyd/traits/.*)" +
		")$",
)

func StemWalkShouldMatch(path string, info fs.FileInfo, basePath string) (bool, error) {
	var relPath, relErr = filepath.Rel(basePath, path)
	if relErr != nil {
		return false, relErr
	}
	matchesPath := RE_STEM_WALK_SHOULD_MATCH.Match([]byte(relPath))
	shouldMatch := matchesPath
	return shouldMatch, nil
}

type StemWalkRequest struct {
	BasePath string
	OnMatch  func(path string, info fs.FileInfo, basePath string) error
}

func StemWalk(args StemWalkRequest) error {

	return fs2.BFSWalk(fs2.Walk3Request{
		BasePath:    args.BasePath,
		ShouldCrawl: StemWalkShouldCrawl,
		ShouldMatch: StemWalkShouldMatch,
		OnMatch:     args.OnMatch,
	})
}

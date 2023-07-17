package core

import (
	fs2 "dryad/filesystem"
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

func StemWalkShouldCrawl(context fs2.Walk4Context) (bool, error) {
	// fmt.Println("StemWalkShouldCrawl", context, context.BasePath, context.VPath)

	var relPath, relErr = filepath.Rel(context.BasePath, context.VPath)
	if relErr != nil {
		return false, relErr
	}
	// fmt.Println("StemWalkShouldCrawl relPath", relPath, relErr)

	matchesPath := RE_STEM_WALK_SHOULD_CRAWL.Match([]byte(relPath))
	if !matchesPath {
		// fmt.Println("StemWalkShouldCrawl 1", context.Path, context.BasePath, relPath, false)
		return false, nil
	}

	if context.Info.IsDir() {
		// fmt.Println("StemWalkShouldCrawl 1.5", context.Path, context.BasePath)
		return true, nil
	} else if context.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
		linkTarget, err := os.Readlink(context.Path)
		if err != nil {
			return false, err
		}
		// clean up relative links
		if !filepath.IsAbs(linkTarget) {
			// fmt.Println("StemWalkShouldCrawl cleaning up linkTarget", linkTarget)
			linkTarget = filepath.Clean(filepath.Join(filepath.Dir(context.Path), linkTarget))
			// fmt.Println("StemWalkShouldCrawl cleaning up linkTarget 2", linkTarget)
		}

		// fmt.Println("StemWalkShouldCrawl 2.0", context.Path, context.BasePath, linkTarget, filepath.IsAbs(linkTarget))

		isDescendant, err := fileIsDescendant(linkTarget, context.BasePath)
		if err != nil {
			return false, err
		}

		// fmt.Println("StemWalkShouldCrawl 2", context.Path, context.BasePath, linkTarget, !isDescendant)
		return !isDescendant, nil
	} else {

		// fmt.Println("StemWalkShouldCrawl 3", context.Path, context.BasePath, true)
		return true, nil
	}
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

func StemWalkShouldMatch(context fs2.Walk4Context) (bool, error) {
	var relPath, relErr = filepath.Rel(context.BasePath, context.VPath)
	if relErr != nil {
		return false, relErr
	}
	matchesPath := RE_STEM_WALK_SHOULD_MATCH.Match([]byte(relPath))

	shouldMatch := matchesPath
	// fmt.Println("StemWalkShouldMatch", context.Path, shouldMatch)
	return shouldMatch, nil
}

type StemWalkRequest struct {
	BasePath string
	OnMatch  func(context fs2.Walk4Context) error
}

func StemWalk(args StemWalkRequest) error {

	return fs2.BFSWalk2(fs2.Walk4Request{
		Path:        args.BasePath,
		VPath:       args.BasePath,
		BasePath:    args.BasePath,
		ShouldCrawl: StemWalkShouldCrawl,
		ShouldMatch: StemWalkShouldMatch,
		OnMatch:     args.OnMatch,
	})
}

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
		"|(dyd/readme)" +
		"|(dyd/type)" +
		"|(dyd/fingerprint)" +
		"|(dyd/secrets-fingerprint)" +
		"|(dyd/main)" +
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

// should walk

// - if the vpath does not match the pattern then no
// - else if the node is a symlink then
// 	- if the node is a symlink pointing to a relative location within the package then no
// 	- else yes
// - else if the node is a directory then yes
// - else if the node is a file then no
// - else error?
func StemWalkShouldCrawl(context fs2.Walk4Context) (bool, error) {
	// fmt.Println("StemWalkShouldCrawl", context.VPath)

	var relPath, relErr = filepath.Rel(context.BasePath, context.VPath)
	if relErr != nil {
		return false, relErr
	}
	matchesPath := RE_STEM_WALK_SHOULD_CRAWL.Match([]byte(relPath))

	if !matchesPath {
		// fmt.Println("StemWalkShouldCrawl 1", context.VPath, false)
		return false, nil
	} else if context.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
		linkTarget, err := os.Readlink(context.Path)
		if err != nil {
			return false, err
		}

		// clean up relative links
		if !filepath.IsAbs(linkTarget) {
			linkTarget = filepath.Clean(filepath.Join(filepath.Dir(context.Path), linkTarget))
		}

		isDescendant, err := fileIsDescendant(linkTarget, context.BasePath)
		if err != nil {
			return false, err
		}

		// fmt.Println("StemWalkShouldCrawl 2", context.VPath, context.Path, linkTarget, isDescendant)
		return !isDescendant, nil
	} else if context.Info.IsDir() {
		// fmt.Println("StemWalkShouldCrawl 3", context.VPath, true)
		return true, nil
	} else {
		// fmt.Println("StemWalkShouldCrawl 4", context.VPath, false)
		return false, nil
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

// should match
// - if the vpath does not match the pattern then no,
// - else if the node is a symlink then
// 	- if the node is a symlink pointing to a relative location within the package then yes,
// 	- else no
// - else if the node is a directory then no,
// - else if the node is a file then yes,
// - else error?
func StemWalkShouldMatch(context fs2.Walk4Context) (bool, error) {
	// fmt.Println("StemWalkShouldMatch", context.VPath)

	var relPath, relErr = filepath.Rel(context.BasePath, context.VPath)
	if relErr != nil {
		return false, relErr
	}
	matchesPath := RE_STEM_WALK_SHOULD_MATCH.Match([]byte(relPath))

	if !matchesPath {
		// fmt.Println("StemWalkShouldMatch 1", context.VPath, false)
		return false, nil
	} else if context.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
		linkTarget, err := os.Readlink(context.Path)
		if err != nil {
			return false, err
		}

		// clean up relative links
		if !filepath.IsAbs(linkTarget) {
			linkTarget = filepath.Clean(filepath.Join(filepath.Dir(context.Path), linkTarget))
		}

		isDescendant, err := fileIsDescendant(linkTarget, context.BasePath)
		if err != nil {
			return false, err
		}

		// fmt.Println("StemWalkShouldMatch 2", context.VPath, context.Path, linkTarget, isDescendant)
		return isDescendant, nil
	} else if context.Info.IsDir() {
		// fmt.Println("StemWalkShouldMatch 3", context.VPath, false)
		return false, nil
	} else {
		// fmt.Println("StemWalkShouldMatch 4", context.VPath, true)
		return true, nil
	}

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

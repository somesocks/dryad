package fs2

import (
	"io/fs"
	"path/filepath"
	"regexp"
)

var RE_WALK_DEFAULT_CRAWL_INCLUDE, _ = regexp.Compile("^.*$")
var RE_WALK_DEFAULT_CRAWL_EXCLUDE, _ = regexp.Compile("^$")

var RE_WALK_DEFAULT_MATCH_INCLUDE, _ = regexp.Compile("^.*$")
var RE_WALK_DEFAULT_MATCH_EXCLUDE, _ = regexp.Compile("^$")

type ReWalkArgs struct {
	BasePath     string
	CrawlInclude *regexp.Regexp
	CrawlExclude *regexp.Regexp
	MatchInclude *regexp.Regexp
	MatchExclude *regexp.Regexp
	OnMatch      func(path string, info fs.FileInfo) error
	OnError      func(err error, path string, info fs.FileInfo) error
}

func ReWalk(args ReWalkArgs) error {
	if args.CrawlInclude == nil {
		args.CrawlInclude = RE_WALK_DEFAULT_CRAWL_INCLUDE
	}
	if args.CrawlExclude == nil {
		args.CrawlExclude = RE_WALK_DEFAULT_CRAWL_EXCLUDE
	}
	if args.MatchInclude == nil {
		args.MatchInclude = RE_WALK_DEFAULT_MATCH_INCLUDE
	}
	if args.MatchExclude == nil {
		args.MatchExclude = RE_WALK_DEFAULT_MATCH_EXCLUDE
	}

	var walkRequest = WalkRequest{
		BasePath: args.BasePath,
		CrawlInclude: func(path string, info fs.FileInfo) (bool, error) {
			var relPath, relErr = filepath.Rel(args.BasePath, path)
			if relErr != nil {
				return false, relErr
			}
			return args.CrawlInclude.Match([]byte(relPath)), nil
		},
		CrawlExclude: func(path string, info fs.FileInfo) (bool, error) {
			var relPath, relErr = filepath.Rel(args.BasePath, path)
			if relErr != nil {
				return false, relErr
			}
			return args.CrawlExclude.Match([]byte(relPath)), nil
		},
		MatchInclude: func(path string, info fs.FileInfo) (bool, error) {
			var relPath, relErr = filepath.Rel(args.BasePath, path)
			if relErr != nil {
				return false, relErr
			}
			res := args.MatchInclude.Match([]byte(relPath))
			// fmt.Println("rewalk MatchInclude", path, res)
			return res, nil
		},
		MatchExclude: func(path string, info fs.FileInfo) (bool, error) {
			var relPath, relErr = filepath.Rel(args.BasePath, path)
			if relErr != nil {
				return false, relErr
			}
			return args.MatchExclude.Match([]byte(relPath)), nil
		},
		OnMatch: func(path string, info fs.FileInfo) error {
			return args.OnMatch(path, info)
		},
		OnError: args.OnError,
	}

	return Walk(walkRequest)

}

package fs2

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

var DEFAULT_CRAWL_ALLOW, _ = regexp.Compile("^.*$")
var DEFAULT_CRAWL_DENY, _ = regexp.Compile("^$")

var DEFAULT_MATCH_ALLOW, _ = regexp.Compile("^.*$")
var DEFAULT_MATCH_DENY, _ = regexp.Compile("^$")

func _reWalk(filename string, linkDirname string, walkFn filepath.WalkFunc) error {
	symWalkFunc := func(path string, info os.FileInfo, err error) error {

		if fname, err := filepath.Rel(filename, path); err == nil {
			path = filepath.Join(linkDirname, fname)
		} else {
			return err
		}

		if err == nil && info.Mode()&os.ModeSymlink == os.ModeSymlink {
			finalPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
			info, err := os.Lstat(finalPath)
			if err != nil {
				return walkFn(path, info, err)
			}
			if info.IsDir() {
				return _reWalk(finalPath, path, walkFn)
			}
		}

		return walkFn(path, info, err)
	}
	return filepath.Walk(filename, symWalkFunc)
}

type ReWalkArgs struct {
	BasePath   string
	CrawlAllow *regexp.Regexp
	CrawlDeny  *regexp.Regexp
	MatchAllow *regexp.Regexp
	MatchDeny  *regexp.Regexp
	OnMatch    filepath.WalkFunc
}

// func ReWalk(args ReWalkArgs) error {
// 	if args.CrawlAllow == nil {
// 		args.CrawlAllow = DEFAULT_CRAWL_ALLOW
// 	}
// 	if args.CrawlDeny == nil {
// 		args.CrawlDeny = DEFAULT_CRAWL_DENY
// 	}
// 	if args.MatchAllow == nil {
// 		args.MatchAllow = DEFAULT_MATCH_ALLOW
// 	}
// 	if args.MatchDeny == nil {
// 		args.MatchDeny = DEFAULT_MATCH_DENY
// 	}

// 	err := _reWalk(args.BasePath, args.BasePath, func(path string, info fs.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		var relPath, relErr = filepath.Rel(args.BasePath, path)
// 		if relErr != nil {
// 			return relErr
// 		}

// 		var matchAllow = args.MatchAllow.MatchString(relPath)
// 		var matchDeny = args.MatchDeny.MatchString(relPath)
// 		if matchAllow && !matchDeny {
// 			var result = args.OnMatch(path, info, err)
// 			if result != nil {
// 				return result
// 			}
// 		}

// 		if info.IsDir() {
// 			var crawlAllow = args.CrawlAllow.MatchString(relPath)
// 			var crawlDeny = args.CrawlDeny.MatchString(relPath)
// 			if crawlAllow && !crawlDeny {
// 				return nil
// 			} else {
// 				return filepath.SkipDir
// 			}
// 		}

// 		return nil
// 	})
// 	return err
// }

func ReWalk(args ReWalkArgs) error {
	if args.CrawlAllow == nil {
		args.CrawlAllow = DEFAULT_CRAWL_ALLOW
	}
	if args.CrawlDeny == nil {
		args.CrawlDeny = DEFAULT_CRAWL_DENY
	}
	if args.MatchAllow == nil {
		args.MatchAllow = DEFAULT_MATCH_ALLOW
	}
	if args.MatchDeny == nil {
		args.MatchDeny = DEFAULT_MATCH_DENY
	}

	var walkRequest = WalkRequest{
		BasePath: args.BasePath,
		CrawlInclude: func(path string, info fs.FileInfo) (bool, error) {
			var relPath, relErr = filepath.Rel(args.BasePath, path)
			if relErr != nil {
				return false, relErr
			}
			return args.CrawlAllow.Match([]byte(relPath)), nil
		},
		CrawlExclude: func(path string, info fs.FileInfo) (bool, error) {
			var relPath, relErr = filepath.Rel(args.BasePath, path)
			if relErr != nil {
				return false, relErr
			}
			return args.CrawlDeny.Match([]byte(relPath)), nil
		},
		MatchInclude: func(path string, info fs.FileInfo) (bool, error) {
			var relPath, relErr = filepath.Rel(args.BasePath, path)
			if relErr != nil {
				return false, relErr
			}
			return args.MatchAllow.Match([]byte(relPath)), nil
		},
		MatchExclude: func(path string, info fs.FileInfo) (bool, error) {
			var relPath, relErr = filepath.Rel(args.BasePath, path)
			if relErr != nil {
				return false, relErr
			}
			return args.MatchDeny.Match([]byte(relPath)), nil
		},
		OnMatch: func(path string, info fs.FileInfo) error {
			return args.OnMatch(path, info, nil)
		},
	}

	return Walk(walkRequest)
}

package fs2

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var WALK_DEFAULT_CRAWL_INCLUDE = func(path string, info fs.FileInfo) (bool, error) {
	return true, nil
}

var WALK_DEFAULT_CRAWL_EXCLUDE = func(path string, info fs.FileInfo) (bool, error) {
	return false, nil
}

var WALK_DEFAULT_MATCH_INCLUDE = func(path string, info fs.FileInfo) (bool, error) {
	return true, nil
}

var WALK_DEFAULT_MATCH_EXCLUDE = func(path string, info fs.FileInfo) (bool, error) {
	return false, nil
}

var WALK_DEFAULT_ON_MATCH = func(path string, info fs.FileInfo) error {
	return nil
}

var WALK_DEFAULT_ON_ERROR = func(err error, path string, info fs.FileInfo) error {
	return err
}

type Walk2Request struct {
	BasePath     string
	CrawlInclude func(path string, info fs.FileInfo) (bool, error)
	CrawlExclude func(path string, info fs.FileInfo) (bool, error)
	MatchInclude func(path string, info fs.FileInfo) (bool, error)
	MatchExclude func(path string, info fs.FileInfo) (bool, error)
	OnMatch      func(path string, info fs.FileInfo) error
	OnError      func(err error, path string, info fs.FileInfo) error
}

func _walk2(context Walk2Request, path string) error {
	// fmt.Println("_walk2 1", path)
	var err error
	var info fs.FileInfo

	info, err = os.Lstat(path)
	if err != nil {
		err = context.OnError(err, path, info)
		if err != nil {
			return err
		}
	}
	// fmt.Println("_walk2 2", path)

	var matchInclude bool
	matchInclude, err = context.MatchInclude(path, info)
	if err != nil {
		err = context.OnError(err, path, info)
		if err != nil {
			return err
		}
	}

	// fmt.Println("_walk2 3", path)
	var matchExclude bool
	matchExclude, err = context.MatchExclude(path, info)
	if err != nil {
		err = context.OnError(err, path, info)
		if err != nil {
			return err
		}
	}

	// fmt.Println("_walk2 4", path)
	if matchInclude && !matchExclude {
		err = context.OnMatch(path, info)
		if err != nil {
			err = context.OnError(err, path, info)
			if err != nil {
				return err
			}
		}
	}

	// info could still be nil here because of error swallowing in OnError
	if info != nil && info.Mode()&os.ModeSymlink == os.ModeSymlink {
		var linkPath string
		linkPath, err = os.Readlink(path)
		if err != nil {
			err = context.OnError(err, path, info)
			if err != nil {
				return err
			}
		}
		if !filepath.IsAbs(linkPath) {
			linkPath = filepath.Join(
				filepath.Dir(path),
				linkPath,
			)
		}

		err = _walk2(context, linkPath)
		if err != nil {
			err = context.OnError(err, path, info)
			if err != nil {
				return err
			}
		}
	} else if info != nil && info.IsDir() {
		var crawlInclude bool
		crawlInclude, err = context.CrawlInclude(path, info)
		if err != nil {
			err = context.OnError(err, path, info)
			if err != nil {
				return err
			}
		}

		var crawlExclude bool
		crawlExclude, err = context.CrawlExclude(path, info)
		if err != nil {
			err = context.OnError(err, path, info)
			if err != nil {
				return err
			}
		}

		if crawlInclude && !crawlExclude {
			var dir *os.File
			dir, err = os.Open(path)
			if err != nil {
				err = context.OnError(err, path, info)
				if err != nil {
					return err
				}
			}
			defer dir.Close()

			var entries []fs.DirEntry

			entries, err = dir.ReadDir(100)
			if err != nil && err != io.EOF {
				err = context.OnError(err, path, info)
				if err != nil {
					return err
				}
			}

			for len(entries) > 0 {
				for _, entry := range entries {
					err = _walk2(context, filepath.Join(path, entry.Name()))
					if err != nil {
						err = context.OnError(err, path, info)
						if err != nil {
							return err
						}
					}
				}

				entries, err = dir.ReadDir(100)
				if err != nil && err != io.EOF {
					err = context.OnError(err, path, info)
					if err != nil {
						return err
					}
				}
			}
		}

	}

	return nil
}

func Walk2(request Walk2Request) error {
	if request.CrawlInclude == nil {
		request.CrawlInclude = WALK_DEFAULT_CRAWL_INCLUDE
	}

	if request.CrawlExclude == nil {
		request.CrawlExclude = WALK_DEFAULT_CRAWL_EXCLUDE
	}

	if request.MatchInclude == nil {
		request.MatchInclude = WALK_DEFAULT_MATCH_INCLUDE
	}

	if request.MatchExclude == nil {
		request.MatchExclude = WALK_DEFAULT_MATCH_EXCLUDE
	}

	if request.OnMatch == nil {
		request.OnMatch = WALK_DEFAULT_ON_MATCH
	}

	if request.OnError == nil {
		request.OnError = WALK_DEFAULT_ON_ERROR
	}

	return _walk2(request, request.BasePath)
}

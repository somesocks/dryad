package fs2

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var DEFAULT_CRAWL_INCLUDE = func(path string, info fs.FileInfo) (bool, error) {
	return true, nil
}

var DEFAULT_CRAWL_EXCLUDE = func(path string, info fs.FileInfo) (bool, error) {
	return false, nil
}

var DEFAULT_MATCH_INCLUDE = func(path string, info fs.FileInfo) (bool, error) {
	return true, nil
}

var DEFAULT_MATCH_EXCLUDE = func(path string, info fs.FileInfo) (bool, error) {
	return false, nil
}

type WalkRequest struct {
	BasePath     string
	CrawlInclude func(path string, info fs.FileInfo) (bool, error)
	CrawlExclude func(path string, info fs.FileInfo) (bool, error)
	MatchInclude func(path string, info fs.FileInfo) (bool, error)
	MatchExclude func(path string, info fs.FileInfo) (bool, error)
	OnMatch      func(path string, info fs.FileInfo) error
}

func _walk(context WalkRequest, path string) error {
	var err error
	var realPath string
	var info fs.FileInfo

	realPath, err = filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}

	info, err = os.Lstat(realPath)
	if err != nil {
		return err
	}

	var matchInclude bool
	matchInclude, err = context.MatchInclude(path, info)
	if err != nil {
		return err
	}

	var matchExclude bool
	matchExclude, err = context.MatchExclude(path, info)
	if err != nil {
		return err
	}

	if matchInclude && !matchExclude {
		err = context.OnMatch(path, info)
		if err != nil {
			return err
		}
	}

	if info.IsDir() {
		var crawlInclude bool
		crawlInclude, err = context.CrawlInclude(path, info)
		if err != nil {
			return err
		}

		var crawlExclude bool
		crawlExclude, err = context.CrawlExclude(path, info)
		if err != nil {
			return err
		}

		if crawlInclude && !crawlExclude {
			var dir *os.File
			dir, err = os.Open(realPath)
			if err != nil {
				return err
			}
			defer dir.Close()

			var entries []fs.DirEntry

			entries, err = dir.ReadDir(100)
			if err != nil && err != io.EOF {
				return err
			}

			for len(entries) > 0 {
				for _, entry := range entries {
					err = _walk(context, filepath.Join(path, entry.Name()))
					if err != nil {
						return err
					}
				}

				entries, err = dir.ReadDir(100)
				if err != nil && err != io.EOF {
					return err
				}
			}
		}

	}

	return nil
}

func Walk(request WalkRequest) error {
	if request.CrawlInclude == nil {
		request.CrawlInclude = DEFAULT_CRAWL_INCLUDE
	}

	if request.CrawlExclude == nil {
		request.CrawlExclude = DEFAULT_CRAWL_EXCLUDE
	}

	if request.MatchInclude == nil {
		request.MatchInclude = DEFAULT_MATCH_INCLUDE
	}

	if request.MatchExclude == nil {
		request.MatchExclude = DEFAULT_MATCH_EXCLUDE
	}

	return _walk(request, request.BasePath)
}

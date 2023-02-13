package core

import (
	fs2 "dryad/filesystem"
	"io/fs"
)

func ScopesWalk(path string, walkFn func(path string, info fs.FileInfo) error) error {
	var scopesPath, err = ScopesPath(path)
	if err != nil {
		return err
	}

	// only crawl the first directory level
	var _scopeCrawlExclude = func(path string, info fs.FileInfo) (bool, error) {
		return path != scopesPath, nil
	}

	// directories in the scopes dir are scopes, but not the inital dir
	var _scopeMatchInclude = func(path string, info fs.FileInfo) (bool, error) {
		return info.IsDir() && path != scopesPath, nil
	}

	err = fs2.Walk(fs2.WalkRequest{
		BasePath:     scopesPath,
		CrawlExclude: _scopeCrawlExclude,
		MatchInclude: _scopeMatchInclude,
		OnMatch:      walkFn,
	})
	if err != nil {
		return err
	}

	return nil
}

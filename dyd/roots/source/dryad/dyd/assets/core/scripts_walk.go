package core

import (
	fs2 "dryad/filesystem"
	"fmt"
	"io/fs"
	"regexp"
)

var _SCRIPTS_WALK_CRAWL_INCLUDE, _ = regexp.Compile(`^.*$`)

var _SCRIPTS_WALK_CRAWL_EXCLUDE, _ = regexp.Compile(`^$`)

var _SCRIPTS_WALK_MATCH_INCLUDE, _ = regexp.Compile(`^script-run-.*$`)

var _SCRIPTS_WALK_MATCH_EXCLUDE, _ = regexp.Compile(`^.*\.oneline$`)

type ScriptsWalkRequest struct {
	BasePath string
	Scope    string
	OnMatch  func(path string, info fs.FileInfo) error
}

func ScriptsWalk(request ScriptsWalkRequest) error {
	scopePath, err := ScopePath(request.BasePath, request.Scope)
	if err != nil {
		return err
	}
	scopeExists, err := fileExists(scopePath)
	if err != nil {
		return err
	}
	if !scopeExists {
		return fmt.Errorf("scope %s not found", request.Scope)
	}

	return fs2.ReWalk(fs2.ReWalkArgs{
		BasePath:     scopePath,
		CrawlInclude: _SCRIPTS_WALK_CRAWL_INCLUDE,
		CrawlExclude: _SCRIPTS_WALK_CRAWL_EXCLUDE,
		MatchInclude: _SCRIPTS_WALK_MATCH_INCLUDE,
		MatchExclude: _SCRIPTS_WALK_MATCH_EXCLUDE,
		OnMatch:      request.OnMatch,
	})

}

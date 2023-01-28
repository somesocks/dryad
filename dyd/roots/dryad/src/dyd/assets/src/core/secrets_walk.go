package core

import (
	fs2 "dryad/filesystem"
	"io/fs"
	"regexp"
)

var _SECRETS_WALK_CRAWL_INCLUDE, _ = regexp.Compile(`^.*$`)

var _SECRETS_WALK_CRAWL_EXCLUDE, _ = regexp.Compile(`^$`)

var _SECRETS_WALK_MATCH_INCLUDE, _ = regexp.Compile(`^.*$`)

var _SECRETS_WALK_MATCH_EXCLUDE, _ = regexp.Compile(`^$`)

type SecretsWalkArgs struct {
	BasePath string
	OnMatch  func(path string, info fs.FileInfo) error
}

func SecretsWalk(args SecretsWalkArgs) error {
	return fs2.ReWalk(fs2.ReWalkArgs{
		BasePath:     args.BasePath,
		CrawlInclude: _SECRETS_WALK_CRAWL_INCLUDE,
		CrawlExclude: _SECRETS_WALK_CRAWL_EXCLUDE,
		MatchInclude: _SECRETS_WALK_MATCH_INCLUDE,
		MatchExclude: _SECRETS_WALK_MATCH_EXCLUDE,
		OnMatch:      args.OnMatch,
	})
}

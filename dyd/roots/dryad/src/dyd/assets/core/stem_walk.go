package core

import (
	fs2 "dryad/filesystem"
	"io/fs"
	"regexp"
)

var STEM_WALK_CRAWL_INCLUDE, _ = regexp.Compile(`^((\.)|(dyd)|(dyd/path)|(dyd/assets)|(dyd/assets/.*)|(dyd/traits)|(dyd/traits/.*)|(dyd/stems)|(dyd/stems/[^/]*)|(dyd/stems/.*/dyd)|(dyd/stems/.*/dyd/traits(/.*)?))$`)

var STEM_WALK_CRAWL_EXCLUDE, _ = regexp.Compile(`^$`)

var STEM_WALK_MATCH_INCLUDE, _ = regexp.Compile(`^((dyd/path/.*)|(dyd/assets/.*)|(dyd/readme)|(dyd/fingerprint)|(dyd/secrets-fingerprint)|(dyd/main)|(dyd/stems/.*/dyd/fingerprint)|(dyd/stems/.*/dyd/traits/.*)|(dyd/traits/.*))$`)

var STEM_WALK_MATCH_EXCLUDE, _ = regexp.Compile(`^$`)

type StemWalkArgs struct {
	BasePath     string
	CrawlInclude *regexp.Regexp
	CrawlExclude *regexp.Regexp
	MatchInclude *regexp.Regexp
	MatchExclude *regexp.Regexp
	OnMatch      func(path string, info fs.FileInfo) error
}

func StemWalk(args StemWalkArgs) error {
	if args.CrawlInclude == nil {
		args.CrawlInclude = STEM_WALK_CRAWL_INCLUDE
	}

	if args.CrawlExclude == nil {
		args.CrawlExclude = STEM_WALK_CRAWL_EXCLUDE
	}

	if args.MatchInclude == nil {
		args.MatchInclude = STEM_WALK_MATCH_INCLUDE
	}

	if args.MatchExclude == nil {
		args.MatchExclude = STEM_WALK_MATCH_EXCLUDE
	}

	return fs2.ReWalk(fs2.ReWalkArgs{
		BasePath:     args.BasePath,
		CrawlInclude: args.CrawlInclude,
		CrawlExclude: args.CrawlExclude,
		MatchInclude: args.MatchInclude,
		MatchExclude: args.MatchExclude,
		OnMatch:      args.OnMatch,
	})
}

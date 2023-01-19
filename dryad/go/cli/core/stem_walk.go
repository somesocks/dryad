package core

import (
	"dryad/filesystem"
	"path/filepath"
	"regexp"
)

var STEM_CRAWL_ALLOW, _ = regexp.Compile(`^((\.)|(dyd)|(dyd/path)|(dyd/assets)|(dyd/assets/.*)|(dyd/traits)|(dyd/traits/.*)|(dyd/stems)|(dyd/stems/[^/]*)|(dyd/stems/.*/dyd)|(dyd/stems/.*/dyd/traits(/.*)?))$`)

var STEM_CRAWL_DENY, _ = regexp.Compile(`^$`)

var STEM_MATCH_ALLOW, _ = regexp.Compile(`^((dyd/path/.*)|(dyd/assets/.*)|(dyd/fingerprint)|(dyd/main)|(dyd/env)|(dyd/stems/.*/dyd/fingerprint)|(dyd/stems/.*/dyd/traits/.*)|(dyd/traits/.*))$`)

var STEM_MATCH_DENY, _ = regexp.Compile(`^$`)

type StemWalkArgs struct {
	BasePath   string
	CrawlAllow *regexp.Regexp
	CrawlDeny  *regexp.Regexp
	MatchAllow *regexp.Regexp
	MatchDeny  *regexp.Regexp
	OnMatch    filepath.WalkFunc
}

func StemWalk(args StemWalkArgs) error {
	if args.CrawlAllow == nil {
		args.CrawlAllow = STEM_CRAWL_ALLOW
	}

	if args.CrawlDeny == nil {
		args.CrawlDeny = STEM_CRAWL_DENY
	}

	if args.MatchAllow == nil {
		args.MatchAllow = STEM_MATCH_ALLOW
	}

	if args.MatchDeny == nil {
		args.MatchDeny = STEM_MATCH_DENY
	}

	return filesystem.ReWalk(filesystem.ReWalkArgs{
		BasePath:   args.BasePath,
		CrawlAllow: args.CrawlAllow,
		CrawlDeny:  args.CrawlDeny,
		MatchAllow: args.MatchAllow,
		MatchDeny:  args.MatchDeny,
		OnMatch:    args.OnMatch,
	})
}

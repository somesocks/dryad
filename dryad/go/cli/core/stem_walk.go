package core

import (
	"dryad/filesystem"
	"path/filepath"
	"regexp"
)

var STEM_DIRS_MATCH, _ = regexp.Compile(`^((\.)|(dyd)|(dyd/path)|(dyd/assets)|(dyd/assets/.*)|(dyd/traits)|(dyd/traits/.*)|(dyd/stems)|(dyd/stems/[^/]*)|(dyd/stems/.*/dyd)|(dyd/stems/.*/dyd/traits(/.*)?))$`)

var STEM_FILES_MATCH, _ = regexp.Compile(`^((dyd/path/.*)|(dyd/assets/.*)|(dyd/fingerprint)|(dyd/main)|(dyd/env)|(dyd/stems/.*/dyd/fingerprint)|(dyd/stems/.*/dyd/traits/.*)|(dyd/traits/.*))$`)

type StemWalkArgs struct {
	BasePath   string
	CrawlAllow *regexp.Regexp
	CrawlDeny  *regexp.Regexp
	MatchAllow *regexp.Regexp
	MatchDeny  *regexp.Regexp
	OnMatch    filepath.WalkFunc
}

func StemWalk(args StemWalkArgs) error {
	return filesystem.ReWalk(filesystem.ReWalkArgs{
		BasePath:   args.BasePath,
		CrawlAllow: args.CrawlAllow,
		CrawlDeny:  args.CrawlDeny,
		MatchAllow: args.MatchAllow,
		MatchDeny:  args.MatchDeny,
		OnMatch:    args.OnMatch,
	})
}

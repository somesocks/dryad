package core

import (
	"dryad/filesystem"
	"path/filepath"
	"regexp"
)

var STEM_DIRS_MATCH, _ = regexp.Compile(`^((\.)|(dyd)|(dyd/path)|(dyd/assets)|(dyd/assets/.*)|(dyd/traits)|(dyd/traits/.*)|(dyd/stems)|(dyd/stems/[^/]*)|(dyd/stems/.*/dyd)|(dyd/stems/.*/dyd/traits(/.*)?))$`)

var STEM_FILES_MATCH, _ = regexp.Compile(`^((dyd/path/.*)|(dyd/assets/.*)|(dyd/fingerprint)|(dyd/main)|(dyd/env)|(dyd/stems/.*/dyd/fingerprint)|(dyd/stems/.*/dyd/traits/.*)|(dyd/traits/.*))$`)

func StemWalk(path string, walkFn filepath.WalkFunc) error {
	return filesystem.ReWalk(filesystem.ReWalkArgs{
		BasePath:    path,
		CrawlFilter: STEM_DIRS_MATCH,
		MatchFilter: STEM_FILES_MATCH,
		OnMatch:     walkFn,
	})
}

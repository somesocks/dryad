package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

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
	BasePath    string
	CrawlFilter *regexp.Regexp
	MatchFilter *regexp.Regexp
	OnMatch     filepath.WalkFunc
}

func ReWalk(args ReWalkArgs) error {

	err := _reWalk(args.BasePath, args.BasePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var relPath, relErr = filepath.Rel(args.BasePath, path)
		if relErr != nil {
			return relErr
		}

		var match = args.MatchFilter.MatchString(relPath)
		if match {
			var result = args.OnMatch(path, info, err)
			if result != nil {
				return result
			}
		}

		if info.IsDir() {
			var crawl = args.CrawlFilter.MatchString(relPath)
			if crawl {
				return nil
			} else {
				return filepath.SkipDir
			}
		}

		return nil
	})
	return err
}

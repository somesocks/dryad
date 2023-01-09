package core

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

var STEM_REGEXP = `^((dyd/assets/.*)|(dyd/fingerprint)|(dyd/main)|(dyd/stems/.*/dyd/fingerprint)|(dyd/stems/.*/dyd/traits/.*)|(dyd/traits/.*))$`

func walk(filename string, linkDirname string, walkFn filepath.WalkFunc) error {
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
				return walk(finalPath, path, walkFn)
			}
		}

		return walkFn(path, info, err)
	}
	return filepath.Walk(filename, symWalkFunc)
}

func StemWalk(path string, walkFn filepath.WalkFunc) error {
	var stem_path, err = StemPath(path)
	// log.Print("stem_path ", stem_path)
	walk(stem_path, stem_path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var relPath, relErr = filepath.Rel(stem_path, path)
		if relErr != nil {
			return relErr
		}

		var re, reErr = regexp.Compile(STEM_REGEXP)
		if reErr != nil {
			return reErr
		}

		var match = re.MatchString(relPath)

		if match {
			return walkFn(path, info, err)
		}

		return nil
	})
	return err
}

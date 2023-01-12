package core

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

var STEM_DIRS_MATCH = `^((\.)|(dyd)|(dyd/assets)|(dyd/assets/.*)|(dyd/traits)|(dyd/traits/.*)|(dyd/stems)|(dyd/stems/[^/]*)|(dyd/stems/.*/dyd)|(dyd/stems/.*/dyd/traits(/.*)?))$`

var STEM_FILES_MATCH = `^((dyd/assets/.*)|(dyd/fingerprint)|(dyd/main)|(dyd/stems/.*/dyd/fingerprint)|(dyd/stems/.*/dyd/traits/.*)|(dyd/traits/.*))$`

func stemWalk(filename string, linkDirname string, walkFn filepath.WalkFunc) error {
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
				return stemWalk(finalPath, path, walkFn)
			}
		}

		return walkFn(path, info, err)
	}
	return filepath.Walk(filename, symWalkFunc)
}

func StemWalk(path string, walkFn filepath.WalkFunc) error {
	var stem_path, err = StemPath(path)
	// log.Print("stem_path ", stem_path)
	err = stemWalk(stem_path, stem_path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			// fmt.Println("stemwalk patherror ", stem_path, " ", path, " ", err)
			return err
		}

		var relPath, relErr = filepath.Rel(stem_path, path)
		if relErr != nil {
			return relErr
		}

		if info.IsDir() {
			var re, reErr = regexp.Compile(STEM_DIRS_MATCH)
			if reErr != nil {
				return reErr
			}

			var match = re.MatchString(relPath)
			// fmt.Println("StemWalk dir match (", path, ") (", relPath, ") (", match, " ", re)

			if match {
				return nil
			} else {
				return filepath.SkipDir
			}
		} else {
			var re, reErr = regexp.Compile(STEM_FILES_MATCH)
			if reErr != nil {
				return reErr
			}

			var match = re.MatchString(relPath)
			// fmt.Println("StemWalk file match ", path, " ", relPath, " ", match)

			if match {
				return walkFn(path, info, err)
			} else {
				return nil
			}
		}

	})
	return err
}

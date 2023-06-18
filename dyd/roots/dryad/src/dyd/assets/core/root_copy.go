package core

import (
	fs2 "dryad/filesystem"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

var _ROOT_COPY_CRAWL_INCLUDE_REGEXP = regexp.MustCompile(
	"^(" +
		"(\\.)" +
		"|(dyd)" +
		"|(dyd/path)" +
		"|(dyd/assets)" +
		"|(dyd/assets/.*)" +
		"|(dyd/traits)" +
		"|(dyd/traits/.*)" +
		"|(dyd/roots)" +
		"|(dyd/roots/.*)" +
		"|(dyd/stems)" +
		"|(dyd/stems/[^/]*)" +
		"|(dyd/stems/.*/dyd)" +
		"|(dyd/stems/.*/dyd/traits(/.*)?)" +
		")$",
)

var _ROOT_COPY_CRAWL_EXCLUDE_REGEXP = regexp.MustCompile(`^$`)

var _ROOT_COPY_MATCH_INCLUDE_REGEXP = regexp.MustCompile(
	"^(" +
		"(\\.)" +
		"|(dyd)" +
		"|(dyd/path)" +
		"|(dyd/path/.*)" +
		"|(dyd/assets)" +
		"|(dyd/assets/.*)" +
		"|(dyd/readme)" +
		"|(dyd/fingerprint)" +
		"|(dyd/root)" +
		"|(dyd/secrets-fingerprint)" +
		"|(dyd/main)" +
		"|(dyd/roots)" +
		"|(dyd/roots/.*)" +
		"|(dyd/stems)" +
		"|(dyd/stems/.*/dyd/fingerprint)" +
		"|(dyd/stems/.*/dyd/traits/.*)" +
		"|(dyd/traits)" +
		"|(dyd/traits/.*)" +
		")$",
)

var _ROOT_COPY_MATCH_EXCLUDE_REGEXP = regexp.MustCompile(`^$`)

func RootCopy(sourcePath string, destPath string) error {

	// normalize the source path
	sourcePath, err := RootPath(sourcePath)
	if err != nil {
		return err
	}

	// normalize the destination path
	destPath, err = filepath.Abs(destPath)
	if err != nil {
		return err
	}

	// temporary workaround until RootsPath is more correct
	gardenPath, err := GardenPath(sourcePath)
	if err != nil {
		return err
	}
	rootsPath, err := RootsPath(gardenPath)
	if err != nil {
		return err
	}

	isWithinRoots, err := fileIsDescendant(destPath, rootsPath)
	if err != nil {
		return err
	}
	if !isWithinRoots {
		return fmt.Errorf("destination path %s is outside of roots", destPath)
	}

	// gardenPath, err := GardenPath(sourcePath)
	// if err != nil {
	// 	return err
	// }

	// don't crawl symlinks
	crawlInclude := func(path string, info fs.FileInfo) (bool, error) {
		var relPath, relErr = filepath.Rel(sourcePath, path)
		if relErr != nil {
			return false, relErr
		}

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return false, nil
		}

		return _ROOT_COPY_CRAWL_INCLUDE_REGEXP.Match([]byte(relPath)), nil
	}

	crawlExclude := func(path string, info fs.FileInfo) (bool, error) {
		var relPath, relErr = filepath.Rel(sourcePath, path)
		if relErr != nil {
			return false, relErr
		}

		return _ROOT_COPY_CRAWL_EXCLUDE_REGEXP.Match([]byte(relPath)), nil
	}

	matchInclude := func(path string, info fs.FileInfo) (bool, error) {
		var relPath, relErr = filepath.Rel(sourcePath, path)
		if relErr != nil {
			return false, relErr
		}

		res := _ROOT_COPY_MATCH_INCLUDE_REGEXP.Match([]byte(relPath))

		// fmt.Println("[debug] root copy match include", path, relPath, res)
		return res, nil
	}

	matchExclude := func(path string, info fs.FileInfo) (bool, error) {
		var relPath, relErr = filepath.Rel(sourcePath, path)
		if relErr != nil {
			return false, relErr
		}

		return _ROOT_COPY_MATCH_EXCLUDE_REGEXP.Match([]byte(relPath)), nil
	}

	onMatch := func(targetSourcePath string, info fs.FileInfo) error {

		targetRelPath, err := filepath.Rel(sourcePath, targetSourcePath)
		if err != nil {
			return err
		}

		targetDestPath := filepath.Join(destPath, targetRelPath)
		targetDestExists, err := fileExists(targetDestPath)
		if err != nil {
			return err
		} else if targetDestExists {
			return fmt.Errorf("error: copy destination %s already exists", targetDestPath)
		}

		if info.IsDir() {

			// for a directory, make a new dir
			// fmt.Println("[debug] root copy dir", targetSourcePath, targetDestPath)

			err = os.MkdirAll(targetDestPath, info.Mode())
			return err

		} else if info.Mode()&os.ModeSymlink == os.ModeSymlink {

			// for a symlink, make a new link resolving to the target
			// fmt.Println("[debug] root copy symlink", targetSourcePath, targetDestPath)

			linkPath, err := os.Readlink(targetSourcePath)
			if err != nil {
				return err
			}

			// convert relative links to an absolute path
			if !filepath.IsAbs(linkPath) {
				linkPath = filepath.Join(
					filepath.Dir(targetSourcePath),
					linkPath,
				)
			}

			// fmt.Println("[debug] root copy symlink linkpath", linkPath)

			linkRelPath, err := filepath.Rel(filepath.Dir(targetDestPath), linkPath)
			if err != nil {
				return err
			}

			err = os.Symlink(linkRelPath, targetDestPath)
			return err

		} else {

			// for a file, copy contents
			// fmt.Println("[debug] root copy file", targetSourcePath, targetDestPath)

			srcFile, err := os.Open(targetSourcePath)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			var destFile *os.File
			destFile, err = os.Create(targetDestPath)
			if err != nil {
				return err
			}
			defer destFile.Close()

			_, err = destFile.ReadFrom(srcFile)
			if err != nil {
				return err
			}

			err = destFile.Chmod(info.Mode())
			if err != nil {
				return err
			}

			err = destFile.Sync()
			return err

		}
	}

	err = fs2.Walk2(fs2.Walk2Request{
		BasePath:     sourcePath,
		CrawlInclude: crawlInclude,
		CrawlExclude: crawlExclude,
		MatchInclude: matchInclude,
		MatchExclude: matchExclude,
		OnMatch:      onMatch,
	})
	return err
}

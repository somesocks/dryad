package core

import (
	fs2 "dryad/filesystem"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func RootReplace(sourcePath string, destPath string) error {

	// normalize the source path
	sourcePath, err := RootPath(sourcePath)
	if err != nil {
		return err
	}

	// normalize the replacement path
	destPath, err = RootPath(destPath)
	if err != nil {
		return err
	}

	rootsPath, err := RootsPath(sourcePath)
	if err != nil {
		return err
	}

	// don't crawl symlinks
	crawlInclude := func(path string, info fs.FileInfo) (bool, error) {
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return false, nil
		} else {
			return true, nil
		}
	}

	// only match symlinks
	matchInclude := func(path string, info fs.FileInfo) (bool, error) {
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return true, nil
		} else {
			return false, nil
		}
	}

	onMatch := func(targetSourcePath string, info fs.FileInfo) error {

		// ignore non-symlinks
		if info.Mode()&os.ModeSymlink != os.ModeSymlink {
			return fmt.Errorf("error: should be symlink")
		}

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

		// ignore links that are not to the source path
		if linkPath != sourcePath {
			return nil
		}

		destLinkPath, err := filepath.Rel(filepath.Dir(targetSourcePath), destPath)
		if err != nil {
			return err
		}

		err = os.Remove(targetSourcePath)
		if err != nil {
			return err
		}

		err = os.Symlink(destLinkPath, targetSourcePath)
		return err

	}

	err = fs2.Walk2(fs2.Walk2Request{
		BasePath:     rootsPath,
		CrawlInclude: crawlInclude,
		MatchInclude: matchInclude,
		OnMatch:      onMatch,
	})
	return err
}

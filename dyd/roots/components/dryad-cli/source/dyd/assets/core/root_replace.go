package core

import (
	fs2 "dryad/filesystem"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type RootReplaceRequest struct {
	Garden *SafeGardenReference
	SourcePath string
	DestPath string
}

func RootReplace(req RootReplaceRequest) error {

	// normalize the source path
	sourcePath, err := RootPath(req.SourcePath, "")
	if err != nil {
		return err
	}

	// normalize the replacement path
	destPath, err := RootPath(req.DestPath, "")
	if err != nil {
		return err
	}

	rootsPath, err := RootsPath(req.Garden)
	if err != nil {
		return err
	}

	// don't crawl symlinks
	crawlInclude := func(path string, info fs.FileInfo) (bool, error) {
		crawl := info.Mode()&os.ModeSymlink != os.ModeSymlink
		// fmt.Println("[debug] root replace crawl include ", path, crawl)
		return crawl, nil
	}

	crawlExclude := func(path string, info fs.FileInfo) (bool, error) {
		return false, nil
	}

	// only match symlinks
	matchInclude := func(path string, info fs.FileInfo) (bool, error) {
		match := info.Mode()&os.ModeSymlink == os.ModeSymlink
		// fmt.Println("[debug] root replace match include ", path, match)
		return match, nil
	}

	matchExclude := func(path string, info fs.FileInfo) (bool, error) {
		return false, nil
	}

	onMatch := func(targetSourcePath string, info fs.FileInfo) error {

		// fmt.Println("[debug] root replace match", targetSourcePath)

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
		CrawlExclude: crawlExclude,
		MatchInclude: matchInclude,
		MatchExclude: matchExclude,
		OnMatch:      onMatch,
	})
	return err
}

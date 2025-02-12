package core

import (
	fs2 "dryad/filesystem"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type RootReplaceRequest struct {
	Source *SafeRootReference
	Dest *SafeRootReference
}

func RootReplace(req RootReplaceRequest) error {
	var sourcePath string = req.Source.BasePath
	var destPath string = req.Dest.BasePath
	var err error

	// check that source and destination are within the same garden
	if req.Source.Garden.BasePath != req.Dest.Garden.BasePath {
		return fmt.Errorf("source and destination roots are not in same garden")
	}

	rootsPath, err := RootsPath(req.Source.Garden)
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

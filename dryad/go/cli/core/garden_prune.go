package core

import (
	"dryad/filesystem"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var GARDEN_PRUNE_STEMS_CRAWL_ALLOW, _ = regexp.Compile(`^((\.)|(stems))$`)
var GARDEN_PRUNE_STEMS_MATCH_ALLOW, _ = regexp.Compile(`^(stems/.*)$`)

var GARDEN_PRUNE_FILES_CRAWL_ALLOW, _ = regexp.Compile(`^((\.)|(files))$`)
var GARDEN_PRUNE_FIlES_MATCH_ALLOW, _ = regexp.Compile(`^(files/.*)$`)

func GardenPrune(gardenPath string) error {

	// truncate the prune operation to a second,
	// to avoid issues with most filesystems with low-resolution timestamps
	currentTime := time.Now().Local().Truncate(time.Second)

	// normalize garden path
	gardenPath, err := GardenPath(gardenPath)
	if err != nil {
		return err
	}

	sproutsPath := filepath.Join(gardenPath, "dyd", "sprouts")

	// we mark both the symlink and the referenced file
	markFile := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return err
		}

		err = os.Chtimes(path, currentTime, currentTime)
		if err != nil {
			return err
		}

		err = os.Chtimes(realPath, currentTime, currentTime)
		if err != nil {
			return err
		}

		return nil
	}

	err = filesystem.ReWalk(filesystem.ReWalkArgs{
		BasePath: sproutsPath,
		OnMatch:  markFile,
	})
	if err != nil {
		return err
	}

	heapPath := filepath.Join(gardenPath, "dyd", "heap")

	sweepFile := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.ModTime().Before(currentTime) {
			err = os.RemoveAll(path)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = filesystem.ReWalk(filesystem.ReWalkArgs{
		BasePath:   heapPath,
		CrawlAllow: GARDEN_PRUNE_STEMS_CRAWL_ALLOW,
		MatchAllow: GARDEN_PRUNE_STEMS_MATCH_ALLOW,
		OnMatch:    sweepFile,
	})
	if err != nil {
		return err
	}

	err = filesystem.ReWalk(filesystem.ReWalkArgs{
		BasePath:   heapPath,
		CrawlAllow: GARDEN_PRUNE_FILES_CRAWL_ALLOW,
		MatchAllow: GARDEN_PRUNE_FIlES_MATCH_ALLOW,
		OnMatch:    sweepFile,
	})
	if err != nil {
		return err
	}

	return nil
}

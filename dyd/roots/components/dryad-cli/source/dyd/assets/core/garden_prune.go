package core

import (
	fs2 "dryad/filesystem"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

var REGEX_GARDEN_PRUNE_STEMS_CRAWL = regexp.MustCompile(`^((\.)|(stems))$`)
var REGEX_GARDEN_PRUNE_STEMS_MATCH = regexp.MustCompile(`^(stems/.*)$`)

var REGEX_GARDEN_PRUNE_FILES_CRAWL = regexp.MustCompile(`^((\.)|(files))$`)
var REGEX_GARDEN_PRUNE_FIlES_MATCH = regexp.MustCompile(`^(files/.*)$`)

var REGEX_GARDEN_PRUNE_DERIVATIONS_CRAWL = regexp.MustCompile(`^((\.)|(derivations))$`)
var REGEX_GARDEN_PRUNE_DERIVATIONS_MATCH = regexp.MustCompile(`^(derivations/.*)$`)

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

	markStatsChecked := 0
	markStatsMarked := 0

	markShouldCrawl := func(node fs2.Walk5Node) (bool, error) {
		zlog.Trace().
			Str("path", node.VPath).
			Msg("garden prune - markShouldCrawl")

			// crawl if we haven't marked already
		return node.Info.ModTime().Before(currentTime), nil
	}

	markShouldMatch := func(node fs2.Walk5Node) (bool, error) {
		markStatsChecked += 1

		zlog.Trace().
			Str("path", node.VPath).
			Msg("garden prune - markShouldMatch")

		// match if we haven't marked already
		return node.Info.ModTime().Before(currentTime), nil
	}

	markOnMatch := func(node fs2.Walk5Node) error {
		markStatsMarked += 1

		zlog.Trace().
			Str("path", node.VPath).
			Msg("garden prune - markOnMatch")

		err = os.Chtimes(node.Path, currentTime, currentTime)
		if err != nil {
			return err
		}
		return nil
	}

	err = fs2.DFSWalk3(
		task.DEFAULT_CONTEXT,
		fs2.Walk5Request{
			Path: sproutsPath,
			VPath: sproutsPath,
			BasePath: sproutsPath,
			ShouldCrawl: markShouldCrawl,
			ShouldMatch: markShouldMatch,
			OnMatch: markOnMatch,
		},
	)
	if err != nil {
		return err
	}

	zlog.Info().
		Int("checked", markStatsChecked).
		Int("marked", markStatsMarked).
		Msg("garden prune - files marked")

	heapPath := filepath.Join(gardenPath, "dyd", "heap")

	sweepStemShouldCrawl := func(path string, info fs.FileInfo, basePath string) (bool, error) {
		var relPath, relErr = filepath.Rel(heapPath, path)
		if relErr != nil {
			return false, relErr
		}
		matchesPath := REGEX_GARDEN_PRUNE_STEMS_CRAWL.Match([]byte(relPath))
		isSymlink := info.Mode()&os.ModeSymlink == os.ModeSymlink
		shouldCrawl := matchesPath && !isSymlink
		// fmt.Println("sweepStemShouldCrawl", path, relPath, shouldCrawl)
		return shouldCrawl, nil
	}

	sweepStemShouldMatch := func(path string, info fs.FileInfo, basePath string) (bool, error) {
		var relPath, relErr = filepath.Rel(heapPath, path)
		if relErr != nil {
			return false, relErr
		}
		shouldMatch := REGEX_GARDEN_PRUNE_STEMS_MATCH.Match([]byte(relPath))
		return shouldMatch, nil
	}

	sweepStemStatsCheck := 0
	sweepStemStatsCount := 0

	sweepStem := func(path string, info fs.FileInfo, basePath string) error {
		sweepStemStatsCheck += 1

		if info.ModTime().Before(currentTime) {
			err = fs2.RemoveAll(path)
			if err != nil {
				return err
			}

			sweepStemStatsCount += 1
		}

		return nil
	}

	err = fs2.BFSWalk(fs2.Walk3Request{
		BasePath:    heapPath,
		ShouldCrawl: sweepStemShouldCrawl,
		ShouldMatch: sweepStemShouldMatch,
		OnMatch:     sweepStem,
	})
	if err != nil {
		return err
	}

	zlog.Info().
		Int("checked", sweepStemStatsCheck).
		Int("swept", sweepStemStatsCount).
		Msg("garden prune - stems swept")



	sweepDerivationStatsCheck := 0
	sweepDerivationStatsCount := 0	

	sweepDerivationsShouldCrawl := func(path string, info fs.FileInfo, basePath string) (bool, error) {
		relPath, relErr := filepath.Rel(heapPath, path)
		if relErr != nil {
			return false, relErr
		}
		matchesPath := REGEX_GARDEN_PRUNE_DERIVATIONS_CRAWL.Match([]byte(relPath))
		shouldCrawl := matchesPath
		return shouldCrawl, nil
	}

	sweepDerivationsShouldMatch := func(path string, info fs.FileInfo, basePath string) (bool, error) {
		sweepDerivationStatsCheck += 1

		var relPath, relErr = filepath.Rel(heapPath, path)
		if relErr != nil {
			return false, relErr
		}
		matchesPath := REGEX_GARDEN_PRUNE_DERIVATIONS_MATCH.Match([]byte(relPath))

		_, err := os.Stat(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return false, err
		}

		isBroken := err != nil

		shouldMatch := matchesPath && isBroken
		return shouldMatch, nil
	}

	sweepDerivation := func(path string, info fs.FileInfo, basePath string) error {
		sweepDerivationStatsCount += 1
		return os.Remove(path)
	}

	err = fs2.DFSWalk(fs2.Walk3Request{
		BasePath:    heapPath,
		ShouldCrawl: sweepDerivationsShouldCrawl,
		ShouldMatch: sweepDerivationsShouldMatch,
		OnMatch:     sweepDerivation,
	})
	if err != nil {
		return err
	}

	zlog.Info().
		Int("checked", sweepDerivationStatsCheck).
		Int("swept", sweepDerivationStatsCount).
		Msg("garden prune - derivations swept")



	sweepFileStatsCheck := 0
	sweepFileStatsCount := 0	
		
	sweepFileShouldCrawl := func(path string, info fs.FileInfo, basePath string) (bool, error) {
		var relPath, relErr = filepath.Rel(heapPath, path)
		if relErr != nil {
			return false, relErr
		}
		matchesPath := REGEX_GARDEN_PRUNE_FILES_CRAWL.Match([]byte(relPath))
		isSymlink := info.Mode()&os.ModeSymlink == os.ModeSymlink
		shouldCrawl := matchesPath && !isSymlink
		// fmt.Println("sweepStemShouldCrawl", path, relPath, shouldCrawl)
		return shouldCrawl, nil
	}

	sweepFilesShouldMatch := func(path string, info fs.FileInfo, basePath string) (bool, error) {
		sweepFileStatsCheck += 1

		var relPath, relErr = filepath.Rel(heapPath, path)
		if relErr != nil {
			return false, relErr
		}
		shouldMatch := REGEX_GARDEN_PRUNE_FIlES_MATCH.Match([]byte(relPath))
		return shouldMatch, nil
	}

	sweepFile := func(path string, info fs.FileInfo, basePath string) error {
		if info.ModTime().Before(currentTime) {
			parentInfo, err := os.Lstat(filepath.Dir(path))
			if err != nil {
				return err
			}

			if parentInfo.Mode()&0o200 != 0o200 {
				err := os.Chmod(filepath.Dir(path), parentInfo.Mode()|0o200)
				if err != nil {
					return err
				}
			}

			err = os.Remove(path)
			if err != nil {
				return err
			}

			sweepFileStatsCount += 1	
		}

		return nil
	}

	err = fs2.DFSWalk(fs2.Walk3Request{
		BasePath:    heapPath,
		ShouldCrawl: sweepFileShouldCrawl,
		ShouldMatch: sweepFilesShouldMatch,
		OnMatch:     sweepFile,
	})
	if err != nil {
		return err
	}

	zlog.Info().
		Int("checked", sweepFileStatsCheck).
		Int("swept", sweepFileStatsCount).
		Msg("garden prune - files swept")

	return nil
}

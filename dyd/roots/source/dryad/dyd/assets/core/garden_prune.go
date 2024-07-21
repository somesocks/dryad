package core

import (
	fs2 "dryad/filesystem"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

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

	markCount := 0
	markGap := 0

	// we mark both the symlink and the referenced file
	markFile := func(path string, info fs.FileInfo, basePath string) error {
		// fmt.Println("markFile ", path)
		var err error

		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return err
		}

		// set to RWX--X--X temporarily
		err = os.Chmod(path, 0o711)
		if err != nil {
			return err
		}

		err = os.Chtimes(path, currentTime, currentTime)
		if err != nil {
			return err
		}

		// set back to R-X--X--X
		err = os.Chmod(path, 0o511)
		if err != nil {
			return err
		}

		// set to RWX--X--X temporarily
		err = os.Chmod(realPath, 0o711)
		if err != nil {
			return err
		}

		err = os.Chtimes(realPath, currentTime, currentTime)
		if err != nil {
			return err
		}

		// set back to R-X--X--X
		err = os.Chmod(realPath, 0o511)
		if err != nil {
			return err
		}

		markCount += 1
		markGap += 1
		if markGap >= 1000 {
			markGap = 0
			zlog.Info().
				Int("total", markCount).
				Msg("garden prune - marking files to keep")
		}

		return nil
	}

	err = fs2.BFSWalk(fs2.Walk3Request{
		BasePath: sproutsPath,
		OnMatch:  markFile,
	})
	if err != nil {
		return err
	}

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

	sweepStemCount := 0
	sweepStemGap := 0

	sweepStem := func(path string, info fs.FileInfo, basePath string) error {
		if info.ModTime().Before(currentTime) {
			err = fs2.RemoveAll(path)
			if err != nil {
				return err
			}

			sweepStemCount += 1
			sweepStemGap += 1
			if sweepStemGap >= 100 {
				sweepStemGap = 0
				zlog.Info().
					Int("total", sweepStemCount).
					Msg("garden prune - sweeping stems")
			}
	
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

	sweepDerivationCount := 0
	sweepDerivationGap := 0

	sweepDerivation := func(path string, info fs.FileInfo, basePath string) error {
		sweepDerivationCount += 1
		sweepDerivationGap += 1
		if sweepDerivationGap >= 100 {
			sweepDerivationGap = 0
			zlog.Info().
				Int("total", sweepDerivationCount).
				Msg("garden prune - sweeping derivations")
		}	 

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
		var relPath, relErr = filepath.Rel(heapPath, path)
		if relErr != nil {
			return false, relErr
		}
		shouldMatch := REGEX_GARDEN_PRUNE_FIlES_MATCH.Match([]byte(relPath))
		return shouldMatch, nil
	}

	sweepFileCount := 0
	sweepFileGap := 0

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

			sweepFileCount += 1
			sweepFileGap += 1
			if sweepFileGap >= 1000 {
				sweepFileGap = 0
				zlog.Info().
					Int("total", sweepFileCount).
					Msg("garden prune - sweeping files")
			}	 
	
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

	return nil
}

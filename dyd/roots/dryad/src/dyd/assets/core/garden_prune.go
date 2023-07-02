package core

import (
	fs2 "dryad/filesystem"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"
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

	// we mark both the symlink and the referenced file
	markFile := func(path string, info fs.FileInfo) error {
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

	sweepStemShouldCrawl := func(path string, info fs.FileInfo) (bool, error) {
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

	sweepStemShouldMatch := func(path string, info fs.FileInfo) (bool, error) {
		var relPath, relErr = filepath.Rel(heapPath, path)
		if relErr != nil {
			return false, relErr
		}
		shouldMatch := REGEX_GARDEN_PRUNE_STEMS_MATCH.Match([]byte(relPath))
		return shouldMatch, nil
	}

	sweepStem := func(path string, info fs.FileInfo) error {
		if info.ModTime().Before(currentTime) {
			err = fs2.RemoveAll(path)
			if err != nil {
				return err
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

	sweepDerivationsShouldCrawl := func(path string, info fs.FileInfo) (bool, error) {
		relPath, relErr := filepath.Rel(heapPath, path)
		if relErr != nil {
			return false, relErr
		}
		matchesPath := REGEX_GARDEN_PRUNE_DERIVATIONS_CRAWL.Match([]byte(relPath))
		shouldCrawl := matchesPath
		return shouldCrawl, nil
	}

	sweepDerivationsShouldMatch := func(path string, info fs.FileInfo) (bool, error) {
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

	sweepDerivation := func(path string, info fs.FileInfo) error {
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

	sweepFileShouldCrawl := func(path string, info fs.FileInfo) (bool, error) {
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

	sweepFilesShouldMatch := func(path string, info fs.FileInfo) (bool, error) {
		var relPath, relErr = filepath.Rel(heapPath, path)
		if relErr != nil {
			return false, relErr
		}
		shouldMatch := REGEX_GARDEN_PRUNE_FIlES_MATCH.Match([]byte(relPath))
		return shouldMatch, nil
	}

	sweepFile := func(path string, info fs.FileInfo) error {
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

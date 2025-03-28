package core

import (
	fs2 "dryad/filesystem"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"time"
	"strings"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

var REGEX_GARDEN_PRUNE_STEMS_CRAWL = regexp.MustCompile(`^((\.)|(stems))$`)
var REGEX_GARDEN_PRUNE_STEMS_MATCH = regexp.MustCompile(`^(stems/.*)$`)

var REGEX_GARDEN_PRUNE_FILES_CRAWL = regexp.MustCompile(`^((\.)|(files))$`)
var REGEX_GARDEN_PRUNE_FIlES_MATCH = regexp.MustCompile(`^(files/.*)$`)

var REGEX_GARDEN_PRUNE_DERIVATIONS_CRAWL = regexp.MustCompile(`^((\.)|(derivations))$`)
var REGEX_GARDEN_PRUNE_DERIVATIONS_MATCH = regexp.MustCompile(`^(derivations/.*)$`)

type gardenPruneRequest struct {
	Garden *SafeGardenReference
	Snapshot time.Time
}

var gardenPrune_prepareRequest =
	func (ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {

		// truncate the snapshot time to a second,
		// to avoid issues with common filesystems with low-resolution timestamps
		req.Snapshot = req.Snapshot.Truncate(time.Second)
		
		return nil, req
	}

var gardenPrune_mark =
	func (ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {

		sproutsPath := filepath.Join(req.Garden.BasePath, "dyd", "sprouts")

		markStatsChecked := 0
		markStatsMarked := 0

		markShouldCrawl := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
			// crawl if we haven't marked already or the timestamp is newer
			// always crawl the sprouts directory regardless of the timestamp
			var shouldCrawl bool = node.Info.ModTime().Before(req.Snapshot) ||
				strings.HasPrefix(node.Path, sproutsPath)

			zlog.Trace().
				Str("path", node.Path).
				Str("vpath", node.VPath).
				Bool("shouldCrawl", shouldCrawl).
				Time("snapshotTime", req.Snapshot).
				Time("fileTime", node.Info.ModTime()).
				Msg("garden prune - markShouldCrawl")

			return nil, shouldCrawl
		}

		markShouldMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
			// match if we haven't marked already or the timestamp is newer
			// always match the sprouts directory regardless of the timestamp
			var shouldMatch bool = node.Info.ModTime().Before(req.Snapshot) ||
				strings.HasPrefix(node.Path, sproutsPath)

			markStatsChecked += 1

			zlog.Trace().
				Str("path", node.VPath).
				Str("vpath", node.VPath).
				Bool("shouldMatch", shouldMatch).
				Time("snapshotTime", req.Snapshot).
				Time("fileTime", node.Info.ModTime()).
				Msg("garden prune - markShouldMatch")

			return nil, shouldMatch
		}

		markOnMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
			markStatsMarked += 1

			zlog.Trace().
				Str("path", node.VPath).
				Msg("garden prune - markOnMatch")

			var err = os.Chtimes(node.Path, req.Snapshot, req.Snapshot)
			if err != nil {
				return err, nil
			}
			return nil, nil
		}

		var err = fs2.DFSWalk3(
			ctx,
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
			return err, req
		}

		zlog.Info().
			Int("checked", markStatsChecked).
			Int("marked", markStatsMarked).
			Msg("garden prune - files marked")


		return nil, req
	}

var gardenPrune_sweepStems =
	func (ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {
		heapPath := filepath.Join(req.Garden.BasePath, "dyd", "heap")

		sweepStemShouldCrawl := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
			var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
			if relErr != nil {
				return relErr, false
			}
			matchesPath := REGEX_GARDEN_PRUNE_STEMS_CRAWL.Match([]byte(relPath))
			isSymlink := node.Info.Mode()&os.ModeSymlink == os.ModeSymlink
			shouldCrawl := matchesPath && !isSymlink

			zlog.Trace().
				Str("path", node.Path).
				Str("vpath", node.VPath).
				Bool("shouldCrawl", shouldCrawl).
				Msg("GardenPrune/sweepStemShouldCrawl")

			return nil, shouldCrawl
		}

		sweepStemShouldMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
			var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
			if relErr != nil {
				return relErr, false
			}
			shouldMatch := REGEX_GARDEN_PRUNE_STEMS_MATCH.Match([]byte(relPath))

			zlog.Trace().
				Str("path", node.Path).
				Str("vpath", node.VPath).
				Bool("shouldMatch", shouldMatch).
				Msg("GardenPrune/sweepStemShouldMatch")

			return nil, shouldMatch
		}

		sweepStemStatsCheck := 0
		sweepStemStatsCount := 0

		sweepStem := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
			sweepStemStatsCheck += 1

			if node.Info.ModTime().Before(req.Snapshot) {
				var err, _ = fs2.RemoveAll(ctx, node.Path)
				if err != nil {
					return err, nil
				}

				sweepStemStatsCount += 1
			}

			return nil, nil
		}

		var err, _ = fs2.BFSWalk3(
			ctx,
			fs2.Walk5Request{
				BasePath: heapPath,
				Path: heapPath,
				VPath: heapPath,
				ShouldCrawl: sweepStemShouldCrawl,
				ShouldMatch: sweepStemShouldMatch,
				OnMatch:     sweepStem,
			},
		)
		if err != nil {
			return err, req
		}

		zlog.Info().
			Int("checked", sweepStemStatsCheck).
			Int("swept", sweepStemStatsCount).
			Msg("garden prune - stems swept")

		return nil, req
	}

var gardenPrune_sweepDerivations =
	func (ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {
		heapPath := filepath.Join(req.Garden.BasePath, "dyd", "heap")

		sweepDerivationStatsCheck := 0
		sweepDerivationStatsCount := 0	

		sweepDerivationsShouldCrawl := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
			relPath, relErr := filepath.Rel(node.BasePath, node.Path)
			if relErr != nil {
				return relErr, false
			}
			matchesPath := REGEX_GARDEN_PRUNE_DERIVATIONS_CRAWL.Match([]byte(relPath))
			shouldCrawl := matchesPath
			return nil, shouldCrawl
		}

		sweepDerivationsShouldMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
			sweepDerivationStatsCheck += 1

			var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
			if relErr != nil {
				return relErr, false
			}
			matchesPath := REGEX_GARDEN_PRUNE_DERIVATIONS_MATCH.Match([]byte(relPath))

			_, err := os.Stat(node.Path)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err, false
			}

			isBroken := err != nil

			shouldMatch := matchesPath && isBroken
			return nil, shouldMatch
		}

		sweepDerivation := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
			sweepDerivationStatsCount += 1
			return os.Remove(node.Path), nil
		}

		var err = fs2.DFSWalk3(
			ctx,
			fs2.Walk5Request{
				Path:    heapPath,
				VPath:    heapPath,
				BasePath:    heapPath,
				ShouldCrawl: sweepDerivationsShouldCrawl,
				ShouldMatch: sweepDerivationsShouldMatch,
				OnMatch:     sweepDerivation,
			},
		)
		if err != nil {
			return err, req
		}

		zlog.Info().
			Int("checked", sweepDerivationStatsCheck).
			Int("swept", sweepDerivationStatsCount).
			Msg("garden prune - derivations swept")

		return nil, req
	}

var gardenPrune_sweepFiles =
	func (ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {
		heapPath := filepath.Join(req.Garden.BasePath, "dyd", "heap")
		sweepFileStatsCheck := 0
		sweepFileStatsCount := 0	
			
		sweepFileShouldCrawl := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
			var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
			if relErr != nil {
				return relErr, false
			}
			matchesPath := REGEX_GARDEN_PRUNE_FILES_CRAWL.Match([]byte(relPath))
			isSymlink := node.Info.Mode()&os.ModeSymlink == os.ModeSymlink
			shouldCrawl := matchesPath && !isSymlink
			return nil, shouldCrawl
		}

		sweepFilesShouldMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
			sweepFileStatsCheck += 1

			var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
			if relErr != nil {
				return relErr, false
			}
			shouldMatch := REGEX_GARDEN_PRUNE_FIlES_MATCH.Match([]byte(relPath))
			return nil, shouldMatch
		}

		sweepFile := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
			if node.Info.ModTime().Before(req.Snapshot) {
				parentPath := filepath.Dir(node.Path)
				parentInfo, err := os.Lstat(parentPath)
				if err != nil {
					return err, nil
				}

				if parentInfo.Mode()&0o200 != 0o200 {
					err := os.Chmod(parentPath, parentInfo.Mode()|0o200)
					if err != nil {
						return err, nil
					}
				}

				err = os.Remove(node.Path)
				if err != nil {
					return err, nil
				}

				sweepFileStatsCount += 1	
			}

			return nil, nil
		}

		var err = fs2.DFSWalk3(
			ctx,
			fs2.Walk5Request{
				Path:    heapPath,
				VPath:    heapPath,	
				BasePath:    heapPath,
				ShouldCrawl: sweepFileShouldCrawl,
				ShouldMatch: sweepFilesShouldMatch,
				OnMatch:     sweepFile,
			},
		)
		if err != nil {
			return err, req
		}

		zlog.Info().
			Int("checked", sweepFileStatsCheck).
			Int("swept", sweepFileStatsCount).
			Msg("garden prune - files swept")

		return nil, req

	}

var gardenPrune = task.Series6(
	gardenPrune_prepareRequest,
	gardenPrune_mark,
	gardenPrune_sweepStems,
	gardenPrune_sweepDerivations,
	gardenPrune_sweepFiles,
	func (ctx *task.ExecutionContext, req gardenPruneRequest) (error, any) {
		return nil, nil
	},
)

type GardenPruneRequest struct {
	Snapshot time.Time
}

func (sg *SafeGardenReference) Prune(ctx *task.ExecutionContext, req GardenPruneRequest) (error) {
	err, _ := gardenPrune(
		ctx,
		gardenPruneRequest{
			Garden: sg,
			Snapshot: req.Snapshot,
		},
	)
	return err
}
package core

import (
	dydfs "dryad/filesystem"
	"errors"
	"os"
	"io/fs"
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

		markShouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
			// crawl if we haven't marked already or the timestamp is newer
			// always crawl the sprouts directory regardless of the timestamp 
			var shouldCrawl bool = node.Info.ModTime().Before(req.Snapshot) ||
				strings.HasPrefix(node.Path, sproutsPath)

			var isSymlink bool = node.Info.Mode()&os.ModeSymlink == os.ModeSymlink
			if isSymlink {
				var err error
				_, err = os.Stat(node.Path)
				if errors.Is(err, fs.ErrNotExist) {
					shouldCrawl = false
				}
				zlog.Warn().
					Str("path", node.Path).
					Str("vpath", node.VPath).
					Str("action", "garden-prune/mark/should-walk").
					Msg("cannot crawl symlink (broken)")
			}

			zlog.Trace().
				Str("path", node.Path).
				Str("vpath", node.VPath).
				Bool("shouldCrawl", shouldCrawl).
				Time("snapshotTime", req.Snapshot).
				Time("fileTime", node.Info.ModTime()).
				Str("action", "garden-prune/mark/should-walk").
				Msg("")

			return nil, shouldCrawl
		}

		markShouldMatch := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
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
				Str("action", "garden-prune/mark/should-match").
				Msg("")

			return nil, shouldMatch
		}

		markOnMatch := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
			markStatsMarked += 1

			zlog.Trace().
				Str("path", node.VPath).
				Str("action", "garden-prune/mark/on-match").
				Msg("")

			var err = dydfs.Chtimes(node.Path, req.Snapshot, req.Snapshot)
			if err != nil {
				return err, nil
			}
			return nil, nil
		}

		markOnMatch = dydfs.ConditionalWalkAction(markOnMatch, markShouldMatch)

		var err, _ = dydfs.Walk6(
			ctx,
			dydfs.Walk6Request{
				BasePath: sproutsPath,
				Path: sproutsPath,
				VPath: sproutsPath,
				ShouldWalk: markShouldWalk,
				OnPostMatch: markOnMatch,
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

		sweepStemShouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
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
				Str("action", "garden-prune/sweep-stems/should-walk").
				Msg("")

			return nil, shouldCrawl
		}

		sweepStemShouldMatch := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
			var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
			if relErr != nil {
				return relErr, false
			}
			shouldMatch := REGEX_GARDEN_PRUNE_STEMS_MATCH.Match([]byte(relPath))

			zlog.Trace().
				Str("path", node.Path).
				Str("vpath", node.VPath).
				Bool("shouldMatch", shouldMatch).
				Str("action", "garden-prune/sweep-stems/should-match").
				Msg("")

			return nil, shouldMatch
		}

		sweepStemStatsCheck := 0
		sweepStemStatsCount := 0

		sweepStem := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
			sweepStemStatsCheck += 1

			if node.Info.ModTime().Before(req.Snapshot) {
				var err, _ = dydfs.RemoveAll(ctx, node.Path)
				if err != nil {
					return err, nil
				}

				sweepStemStatsCount += 1
			}

			return nil, nil
		}

		sweepStem = dydfs.ConditionalWalkAction(sweepStem, sweepStemShouldMatch)

		var err, _ = dydfs.Walk6(
			ctx,
			dydfs.Walk6Request{
				BasePath: heapPath,
				Path: heapPath,
				VPath: heapPath,
				ShouldWalk: sweepStemShouldWalk,
				OnPreMatch:     sweepStem,
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

		sweepDerivationsShouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
			relPath, relErr := filepath.Rel(node.BasePath, node.Path)
			if relErr != nil {
				return relErr, false
			}
			matchesPath := REGEX_GARDEN_PRUNE_DERIVATIONS_CRAWL.Match([]byte(relPath))
			shouldCrawl := matchesPath
			return nil, shouldCrawl
		}

		sweepDerivationsShouldMatch := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
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

		sweepDerivation := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
			sweepDerivationStatsCount += 1
			return os.Remove(node.Path), nil
		}

		sweepDerivation = dydfs.ConditionalWalkAction(sweepDerivation, sweepDerivationsShouldMatch)

		var err, _ = dydfs.Walk6(
			ctx,
			dydfs.Walk6Request{
				BasePath:    heapPath,
				Path:    heapPath,
				VPath:    heapPath,
				ShouldWalk: sweepDerivationsShouldWalk,
				OnPostMatch:     sweepDerivation,
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
			
		sweepFileShouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
			var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
			if relErr != nil {
				return relErr, false
			}
			matchesPath := REGEX_GARDEN_PRUNE_FILES_CRAWL.Match([]byte(relPath))
			isSymlink := node.Info.Mode()&os.ModeSymlink == os.ModeSymlink
			shouldCrawl := matchesPath && !isSymlink
			return nil, shouldCrawl
		}

		sweepFilesShouldMatch := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
			sweepFileStatsCheck += 1

			var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
			if relErr != nil {
				return relErr, false
			}
			shouldMatch := REGEX_GARDEN_PRUNE_FIlES_MATCH.Match([]byte(relPath))
			return nil, shouldMatch
		}

		sweepFile := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
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

		sweepFile = dydfs.ConditionalWalkAction(sweepFile, sweepFilesShouldMatch)

		var err, _ = dydfs.Walk6(
			ctx,
			dydfs.Walk6Request{
				BasePath:    heapPath,
				Path:    heapPath,
				VPath:    heapPath,	
				ShouldWalk: sweepFileShouldWalk,
				OnPostMatch:     sweepFile,
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
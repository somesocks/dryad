package core

import (
	dydfs "dryad/filesystem"
	"dryad/internal/os"
	"errors"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

var gardenPruneStemLikeLeafPattern = regexp.MustCompile(
	"^[a-z2-7]{2}(?:" + regexp.QuoteMeta(string(filepath.Separator)) + "?[a-z2-7]{2}){12}$",
)

type gardenPruneRequest struct {
	Garden   *SafeGardenReference
	Snapshot time.Time
}

func gardenPruneFingerprintFromVersionPath(versionPath string, objectPath string) (error, string) {
	relativePath, err := filepath.Rel(versionPath, objectPath)
	if err != nil {
		return err, ""
	}

	encoded := strings.ReplaceAll(relativePath, string(filepath.Separator), "")
	fingerprint := fingerprintVersionV2 + "-" + encoded
	err, _, _ = fingerprintParse(fingerprint)
	if err != nil {
		return err, ""
	}
	return nil, fingerprint
}

var gardenPrune_prepareRequest = func(ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {

	// truncate the snapshot time to a second,
	// to avoid issues with common filesystems with low-resolution timestamps
	req.Snapshot = req.Snapshot.Truncate(time.Second)

	return nil, req
}

var gardenPrune_mark = func(ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {

	sproutsPath := filepath.Join(req.Garden.BasePath, "dyd", "sprouts")

	markStatsChecked := 0
	markStatsMarked := 0

	markShouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		isSproutsTraversal := strings.HasPrefix(node.VPath, sproutsPath)

		// crawl if we haven't marked already or the timestamp is newer
		// always crawl the sprouts directory regardless of the timestamp
		var shouldCrawl bool = node.Info.ModTime().Before(req.Snapshot) ||
			isSproutsTraversal

		var isSymlink bool = node.Info.Mode()&os.ModeSymlink == os.ModeSymlink
		if isSymlink {
			var err error
			_, err = os.Stat(node.Path)
			if errors.Is(err, fs.ErrNotExist) {
				shouldCrawl = false

				zlog.Warn().
					Str("path", node.Path).
					Str("vpath", node.VPath).
					Str("action", "garden-prune/mark/should-walk").
					Msg("cannot crawl symlink (broken)")
			}
		}

		zlog.Trace().
			Str("path", node.Path).
			Str("vpath", node.VPath).
			Bool("isSproutsTraversal", isSproutsTraversal).
			Bool("shouldCrawl", shouldCrawl).
			Time("snapshotTime", req.Snapshot).
			Time("fileTime", node.Info.ModTime()).
			Str("action", "garden-prune/mark/should-walk").
			Msg("")

		return nil, shouldCrawl
	}

	markShouldMatch := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		isSproutsTraversal := strings.HasPrefix(node.VPath, sproutsPath)

		// match if we haven't marked already or the timestamp is newer
		// always match the sprouts directory regardless of the timestamp
		var shouldMatch bool = node.Info.ModTime().Before(req.Snapshot) ||
			isSproutsTraversal

		markStatsChecked += 1

		zlog.Trace().
			Str("path", node.Path).
			Str("vpath", node.VPath).
			Bool("isSproutsTraversal", isSproutsTraversal).
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
			BasePath:    sproutsPath,
			Path:        sproutsPath,
			VPath:       sproutsPath,
			ShouldWalk:  markShouldWalk,
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

func gardenPruneDirectoryEmpty(path string) (error, bool) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, true
		}
		return err, false
	}
	return nil, len(entries) == 0
}

func gardenPruneRemoveEmptyDir(path string, rootPath string) (error, bool) {
	if filepath.Clean(path) == filepath.Clean(rootPath) {
		return nil, false
	}

	err, empty := gardenPruneDirectoryEmpty(path)
	if err != nil || !empty {
		return err, false
	}

	err = os.Remove(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err, false
	}

	return nil, true
}

func gardenPruneEnsureParentWritable(path string) error {
	parentPath := filepath.Dir(path)
	parentInfo, err := os.Lstat(parentPath)
	if err != nil {
		return err
	}

	if parentInfo.Mode()&0o200 != 0o200 {
		err = os.Chmod(parentPath, parentInfo.Mode()|0o200)
		if err != nil {
			return err
		}
	}

	return nil
}

func gardenPruneVersionPath(path string) (error, string) {
	exists, err := fileExists(path)
	if err != nil {
		return err, ""
	}
	if !exists {
		return nil, ""
	}
	return nil, path
}

func gardenPruneStemLikeLeaf(path string, versionPath string) (error, bool) {
	relPath, err := filepath.Rel(versionPath, path)
	if err != nil {
		return err, false
	}
	if relPath == "." {
		return nil, false
	}
	return nil, gardenPruneStemLikeLeafPattern.MatchString(relPath)
}

var gardenPrune_sweepStems = func(ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {
	err, stemsVersionPath := gardenPruneVersionPath(heapStemsVersionDir(filepath.Join(req.Garden.BasePath, "dyd", "heap", "stems")))
	if err != nil || stemsVersionPath == "" {
		return err, req
	}

	sweepStemShouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		if !node.Info.IsDir() {
			return nil, false
		}
		err, isLeaf := gardenPruneStemLikeLeaf(node.Path, stemsVersionPath)
		if err != nil {
			return err, false
		}
		zlog.Trace().
			Str("path", node.Path).
			Bool("is_leaf", isLeaf).
			Bool("should_walk", !isLeaf).
			Str("action", "garden-prune/stems/classify").
			Msg("")
		return nil, !isLeaf
	}

	sweepStemStatsCheck := 0
	sweepStemStatsCount := 0

	sweepStemPost := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		if !node.Info.IsDir() {
			return nil, nil
		}
		err, isLeaf := gardenPruneStemLikeLeaf(node.Path, stemsVersionPath)
		if err != nil {
			return err, nil
		}

		if isLeaf {
			isStale := node.Info.ModTime().Before(req.Snapshot)
			zlog.Trace().
				Str("path", node.Path).
				Bool("is_leaf", true).
				Bool("is_stale", isStale).
				Str("action", "garden-prune/stems/delete-start").
				Msg("")

			sweepStemStatsCheck += 1
			if isStale {
				err, _ := dydfs.RemoveAll(ctx, node.Path)
				zlog.Trace().
					Str("path", node.Path).
					Err(err).
					Str("action", "garden-prune/stems/delete-done").
					Msg("")
				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					return err, nil
				}
				sweepStemStatsCount += 1
			}
			return nil, nil
		}

		err, _ = gardenPruneRemoveEmptyDir(node.Path, stemsVersionPath)
		return err, nil
	}

	var errWalk, _ = dydfs.Walk6(
		ctx,
		dydfs.Walk6Request{
			BasePath:    stemsVersionPath,
			Path:        stemsVersionPath,
			VPath:       stemsVersionPath,
			ShouldWalk:  sweepStemShouldWalk,
			OnPostMatch: sweepStemPost,
		},
	)
	if errWalk != nil {
		return errWalk, req
	}

	zlog.Info().
		Int("checked", sweepStemStatsCheck).
		Int("swept", sweepStemStatsCount).
		Msg("garden prune - stems swept")

	return nil, req
}

var gardenPrune_sweepSprouts = func(ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {
	err, sproutsVersionPath := gardenPruneVersionPath(heapSproutsVersionDir(filepath.Join(req.Garden.BasePath, "dyd", "heap", "sprouts")))
	if err != nil || sproutsVersionPath == "" {
		return err, req
	}

	sweepSproutShouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		if !node.Info.IsDir() {
			return nil, false
		}
		err, isLeaf := gardenPruneStemLikeLeaf(node.Path, sproutsVersionPath)
		if err != nil {
			return err, false
		}
		zlog.Trace().
			Str("path", node.Path).
			Bool("is_leaf", isLeaf).
			Bool("should_walk", !isLeaf).
			Str("action", "garden-prune/sprouts/classify").
			Msg("")
		return nil, !isLeaf
	}

	sweepSproutStatsCheck := 0
	sweepSproutStatsCount := 0

	sweepSproutPost := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		if !node.Info.IsDir() {
			return nil, nil
		}
		err, isLeaf := gardenPruneStemLikeLeaf(node.Path, sproutsVersionPath)
		if err != nil {
			return err, nil
		}

		if isLeaf {
			isStale := node.Info.ModTime().Before(req.Snapshot)
			zlog.Trace().
				Str("path", node.Path).
				Bool("is_leaf", true).
				Bool("is_stale", isStale).
				Str("action", "garden-prune/sprouts/delete-start").
				Msg("")

			sweepSproutStatsCheck += 1
			if isStale {
				err, _ := dydfs.RemoveAll(ctx, node.Path)
				zlog.Trace().
					Str("path", node.Path).
					Err(err).
					Str("action", "garden-prune/sprouts/delete-done").
					Msg("")
				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					return err, nil
				}
				sweepSproutStatsCount += 1
			}
			return nil, nil
		}

		err, _ = gardenPruneRemoveEmptyDir(node.Path, sproutsVersionPath)
		return err, nil
	}

	var errWalk, _ = dydfs.Walk6(
		ctx,
		dydfs.Walk6Request{
			BasePath:    sproutsVersionPath,
			Path:        sproutsVersionPath,
			VPath:       sproutsVersionPath,
			ShouldWalk:  sweepSproutShouldWalk,
			OnPostMatch: sweepSproutPost,
		},
	)
	if errWalk != nil {
		return errWalk, req
	}

	zlog.Info().
		Int("checked", sweepSproutStatsCheck).
		Int("swept", sweepSproutStatsCount).
		Msg("garden prune - sprouts swept")

	return nil, req
}

var gardenPrune_sweepDerivations = func(ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {
	heapPath := filepath.Join(req.Garden.BasePath, "dyd", "heap")
	err, derivationsVersionPath := gardenPruneVersionPath(heapDerivationsRootsVersionDir(filepath.Join(heapPath, "derivations")))
	if err != nil || derivationsVersionPath == "" {
		return err, req
	}

	sweepDerivationStatsCheck := 0
	sweepDerivationStatsCount := 0

	sweepDerivationsShouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		return nil, node.Info.IsDir()
	}

	sweepDerivationPost := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		if node.Info.IsDir() {
			err, _ := gardenPruneRemoveEmptyDir(node.Path, derivationsVersionPath)
			return err, nil
		}

		sweepDerivationStatsCheck += 1
		if !node.Info.ModTime().Before(req.Snapshot) {
			return nil, nil
		}
		if !node.Info.Mode().IsRegular() {
			sweepDerivationStatsCount += 1
			return os.RemoveAll(node.Path), nil
		}

		resultFingerprintBytes, err := os.ReadFile(node.Path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, nil
			}
			sweepDerivationStatsCount += 1
			return os.RemoveAll(node.Path), nil
		}
		resultFingerprint := strings.TrimSpace(string(resultFingerprintBytes))
		if resultFingerprint == "" {
			sweepDerivationStatsCount += 1
			return os.Remove(node.Path), nil
		}

		err, sourceFingerprint := gardenPruneFingerprintFromVersionPath(derivationsVersionPath, node.Path)
		if err != nil {
			return err, nil
		}
		err, canonicalDerivationPath := heapDerivationsRootsFingerprintPath(
			ctx,
			req.Garden,
			filepath.Join(heapPath, "derivations"),
			sourceFingerprint,
		)
		if err != nil {
			return err, nil
		}
		if filepath.Clean(node.Path) != filepath.Clean(canonicalDerivationPath) {
			sweepDerivationStatsCount += 1
			return os.Remove(node.Path), nil
		}

		err, resultStemPath := heapStemsFingerprintPath(ctx, req.Garden, filepath.Join(heapPath, "stems"), resultFingerprint)
		if err != nil {
			return err, nil
		}
		_, err = os.Stat(resultStemPath)
		if err == nil {
			return nil, nil
		}
		if errors.Is(err, os.ErrNotExist) {
			sweepDerivationStatsCount += 1
			return os.Remove(node.Path), nil
		}
		return err, nil
	}

	var errWalk, _ = dydfs.Walk6(
		ctx,
		dydfs.Walk6Request{
			BasePath:    derivationsVersionPath,
			Path:        derivationsVersionPath,
			VPath:       derivationsVersionPath,
			ShouldWalk:  sweepDerivationsShouldWalk,
			OnPostMatch: sweepDerivationPost,
		},
	)
	if errWalk != nil {
		return errWalk, req
	}

	zlog.Info().
		Int("checked", sweepDerivationStatsCheck).
		Int("swept", sweepDerivationStatsCount).
		Msg("garden prune - derivations swept")

	return nil, req
}

var gardenPrune_sweepFiles = func(ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {
	err, filesVersionPath := gardenPruneVersionPath(heapFilesVersionDir(filepath.Join(req.Garden.BasePath, "dyd", "heap", "files")))
	if err != nil || filesVersionPath == "" {
		return err, req
	}
	sweepFileStatsCheck := 0
	sweepFileStatsCount := 0

	sweepFileShouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		return nil, node.Info.IsDir()
	}

	sweepFilePost := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		if node.Info.IsDir() {
			err, _ := gardenPruneRemoveEmptyDir(node.Path, filesVersionPath)
			return err, nil
		}

		sweepFileStatsCheck += 1
		if node.Info.Mode().IsRegular() && node.Info.ModTime().Before(req.Snapshot) {
			err := gardenPruneEnsureParentWritable(node.Path)
			if err != nil {
				return err, nil
			}
			err = os.Remove(node.Path)
			if err != nil {
				return err, nil
			}
			sweepFileStatsCount += 1
		}

		return nil, nil
	}

	var errWalk, _ = dydfs.Walk6(
		ctx,
		dydfs.Walk6Request{
			BasePath:    filesVersionPath,
			Path:        filesVersionPath,
			VPath:       filesVersionPath,
			ShouldWalk:  sweepFileShouldWalk,
			OnPostMatch: sweepFilePost,
		},
	)
	if errWalk != nil {
		return errWalk, req
	}

	zlog.Info().
		Int("checked", sweepFileStatsCheck).
		Int("swept", sweepFileStatsCount).
		Msg("garden prune - files swept")

	return nil, req

}

var gardenPrune_sweepSecrets = func(ctx *task.ExecutionContext, req gardenPruneRequest) (error, gardenPruneRequest) {
	err, secretsVersionPath := gardenPruneVersionPath(heapSecretsVersionDir(filepath.Join(req.Garden.BasePath, "dyd", "heap", "secrets")))
	if err != nil || secretsVersionPath == "" {
		return err, req
	}
	sweepSecretStatsCheck := 0
	sweepSecretStatsCount := 0

	sweepSecretShouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		return nil, node.Info.IsDir()
	}

	sweepSecretPost := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		if node.Info.IsDir() {
			err, _ := gardenPruneRemoveEmptyDir(node.Path, secretsVersionPath)
			return err, nil
		}

		sweepSecretStatsCheck += 1
		if node.Info.Mode().IsRegular() && node.Info.ModTime().Before(req.Snapshot) {
			err := gardenPruneEnsureParentWritable(node.Path)
			if err != nil {
				return err, nil
			}
			err = os.Remove(node.Path)
			if err != nil {
				return err, nil
			}
			sweepSecretStatsCount += 1
		}

		return nil, nil
	}

	var errWalk, _ = dydfs.Walk6(
		ctx,
		dydfs.Walk6Request{
			BasePath:    secretsVersionPath,
			Path:        secretsVersionPath,
			VPath:       secretsVersionPath,
			ShouldWalk:  sweepSecretShouldWalk,
			OnPostMatch: sweepSecretPost,
		},
	)
	if errWalk != nil {
		return errWalk, req
	}

	zlog.Info().
		Int("checked", sweepSecretStatsCheck).
		Int("swept", sweepSecretStatsCount).
		Msg("garden prune - secrets swept")

	return nil, req

}

var gardenPrune = task.Series8(
	gardenPrune_prepareRequest,
	gardenPrune_mark,
	gardenPrune_sweepStems,
	gardenPrune_sweepSprouts,
	gardenPrune_sweepDerivations,
	gardenPrune_sweepFiles,
	gardenPrune_sweepSecrets,
	func(ctx *task.ExecutionContext, req gardenPruneRequest) (error, any) {
		return nil, nil
	},
)

type GardenPruneRequest struct {
	Snapshot time.Time
}

func (sg *SafeGardenReference) Prune(ctx *task.ExecutionContext, req GardenPruneRequest) error {
	err, _ := gardenPrune(
		ctx,
		gardenPruneRequest{
			Garden:   sg,
			Snapshot: req.Snapshot,
		},
	)
	return err
}

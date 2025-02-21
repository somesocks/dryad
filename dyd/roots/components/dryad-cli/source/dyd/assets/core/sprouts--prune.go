package core

import (
	dydfs "dryad/filesystem"
	"os"
	"path/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

func sproutsPrune(ctx *task.ExecutionContext, sprouts *SafeSproutsReference) error {
	var roots *SafeRootsReference
	var err error 

	err, roots = sprouts.Garden.Roots().Resolve(ctx)
	if err != nil {
		return err
	}

	// crawl everything that isn't a symlink
	shouldCrawl := func(ctx *task.ExecutionContext, node dydfs.Walk5Node) (error, bool) {
		var shouldCrawl bool = node.Info.Mode()&os.ModeSymlink != os.ModeSymlink

		zlog.Trace().
			Str("path", node.Path).
			Bool("shouldCrawl", shouldCrawl).
			Msg("SproutsPrune.shouldCrawl")
		return nil, shouldCrawl
	}

	// match any path that we should delete
	shouldMatch := func(ctx *task.ExecutionContext, node dydfs.Walk5Node) (error, bool) {

		zlog.Trace().
			Str("path", node.Path).
			Msg("SproutsPrune.shouldMatch")

		relPath, err := filepath.Rel(node.BasePath, node.Path)
		if err != nil {
			return err, false
		}

		rootsEquivalentPath := filepath.Join(roots.BasePath, relPath)

		// if the roots path does not exist, then we should wipe the sprouts path
		rootsPathExists, err := fileExists(rootsEquivalentPath)
		if err != nil {
			return  err, false
		}
		if !rootsPathExists {
			return nil, true
		}

		// TODO: fix the edge case where a sprout path is within a root
		// very unlikely to happen, but possible
		return nil, false
	}

	onMatch := func(ctx *task.ExecutionContext, node dydfs.Walk5Node) (error, any) {
		zlog.Trace().
			Str("path", node.Path).
			Msg("SproutsPrune.onMatch")

		err, _ = dydfs.Remove(ctx, node.Path)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	// NOTE: this needs to run serially until we fix the concurrency issue
	// with parent permissions in onMatch
	err = dydfs.DFSWalk3(
		task.SERIAL_CONTEXT,
		dydfs.Walk5Request{
			Path:        sprouts.BasePath,
			VPath:       sprouts.BasePath,
			BasePath:    sprouts.BasePath,
			ShouldCrawl: shouldCrawl,
			ShouldMatch: shouldMatch,
			OnMatch:     onMatch,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (sprouts *SafeSproutsReference) Prune(
	ctx *task.ExecutionContext,
) (error) {
	err := sproutsPrune(
		ctx,
		sprouts,
	)
	return err
}
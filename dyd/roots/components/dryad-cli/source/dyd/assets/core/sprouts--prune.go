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
	shouldWalk := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		var shouldWalk bool = node.Info.Mode()&os.ModeSymlink != os.ModeSymlink

		zlog.Trace().
			Str("path", node.Path).
			Bool("shouldWalk", shouldWalk).
			Msg("SproutsPrune.shouldWalk")
		return nil, shouldWalk
	}

	// match any path that we should delete
	shouldMatch := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {

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

	onMatch := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		zlog.Trace().
			Str("path", node.Path).
			Msg("SproutsPrune.onMatch")

		err, _ = dydfs.Remove(ctx, node.Path)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	onMatch = dydfs.ConditionalWalkAction(onMatch, shouldMatch)
	
	// NOTE: this needs to run serially until we fix the concurrency issue
	// with parent permissions in onMatch
	err, _ = dydfs.Walk6(
		task.SERIAL_CONTEXT,
		dydfs.Walk6Request{
			BasePath:    sprouts.BasePath,
			Path:        sprouts.BasePath,
			VPath:       sprouts.BasePath,
			ShouldWalk: shouldWalk,
			OnPostMatch:     onMatch,
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
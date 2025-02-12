package core

import (
	fs2 "dryad/filesystem"
	"os"
	"path/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

func SproutsPrune(garden *SafeGardenReference) error {
	zlog.Debug().Msg("pruning sprouts")

	sproutsPath, err := SproutsPath(garden)
	if err != nil {
		return err
	}

	// add a safety check to make sure the sprouts path exists
	// it may not be tracked in a git repo, f.ex.
	sproutsExists, err := fileExists(sproutsPath)
	if err != nil {
		return err
	}
	if !sproutsExists {
		err = os.MkdirAll(sproutsPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	rootsPath, err := RootsPath(garden)
	if err != nil {
		return err
	}

	// crawl everything that isn't a symlink
	shouldCrawl := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		var shouldCrawl bool = node.Info.Mode()&os.ModeSymlink != os.ModeSymlink

		zlog.Trace().
			Str("path", node.Path).
			Bool("shouldCrawl", shouldCrawl).
			Msg("SproutsPrune.shouldCrawl")
		return nil, shouldCrawl
	}

	// match any path that we should delete
	shouldMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {

		zlog.Trace().
			Str("path", node.Path).
			Msg("SproutsPrune.shouldMatch")

		relPath, err := filepath.Rel(node.BasePath, node.Path)
		if err != nil {
			return err, false
		}

		rootsEquivalentPath := filepath.Join(rootsPath, relPath)

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

	onMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
		zlog.Trace().
			Str("path", node.Path).
			Msg("SproutsPrune.onMatch")
		parentPath := filepath.Dir(node.Path)

		// set parent to RWX--X--X temporarily
		err = os.Chmod(parentPath, 0o711)
		if err != nil {
			return err, nil
		}

		err = os.Remove(node.Path)
		if err != nil {
			return err, nil
		}

		// set parent back to R-X--X--X temporarily
		err = os.Chmod(parentPath, 0o511)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	// NOTE: this needs to run serially until we fix the concurrency issue
	// with parent permissions in onMatch
	err = fs2.DFSWalk3(
		task.SERIAL_CONTEXT,
		fs2.Walk5Request{
			Path:        sproutsPath,
			VPath:       sproutsPath,
			BasePath:    sproutsPath,
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

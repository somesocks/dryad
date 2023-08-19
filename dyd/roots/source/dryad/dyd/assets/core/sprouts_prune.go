package core

import (
	fs2 "dryad/filesystem"
	"os"
	"path/filepath"

	log "github.com/rs/zerolog/log"
)

func SproutsPrune(path string) error {
	log.Debug().Msg("pruning sprouts")

	sproutsPath, err := SproutsPath(path)
	log.Trace().
		Str("sproutsPath", sproutsPath).
		Err(err).
		Msg("SproutsPrune.sproutsPath")
	if err != nil {
		return err
	}

	rootsPath, err := RootsPath(path)
	log.Trace().
		Str("rootsPath", rootsPath).
		Err(err).
		Msg("SproutsPrune.rootsPath")
	if err != nil {
		return err
	}

	// crawl everything that isn't a symlink
	shouldCrawl := func(context fs2.Walk4Context) (bool, error) {
		log.Trace().
			Str("path", context.Path).
			Msg("SproutsPrune.shouldCrawl")
		return context.Info.Mode()&os.ModeSymlink != os.ModeSymlink, nil
	}

	// match any path that we should delete
	shouldMatch := func(context fs2.Walk4Context) (bool, error) {
		log.Trace().
			Str("path", context.Path).
			Msg("SproutsPrune.shouldMatch")
		relPath, err := filepath.Rel(context.BasePath, context.Path)
		if err != nil {
			return false, err
		}

		rootsEquivalentPath := filepath.Join(rootsPath, relPath)

		// if the roots path does not exist, then we should wipe the sprouts path
		rootsPathExists, err := fileExists(rootsEquivalentPath)
		if err != nil {
			return false, err
		}
		if !rootsPathExists {
			return true, nil
		}

		// TODO: fix the edge case where a sprout path is within a root
		// very unlikely to happen, but possible
		return false, nil
	}

	onMatch := func(context fs2.Walk4Context) error {
		log.Trace().
			Str("path", context.Path).
			Msg("SproutsPrune.onMatch")
		parentPath := filepath.Dir(context.Path)

		// set parent to RWX--X--X temporarily
		err = os.Chmod(parentPath, 0o711)
		if err != nil {
			return err
		}

		err = os.Remove(context.Path)
		if err != nil {
			return err
		}

		// set parent back to R-X--X--X temporarily
		err = os.Chmod(parentPath, 0o511)
		if err != nil {
			return err
		}

		return nil
	}

	err = fs2.DFSWalk2(fs2.Walk4Request{
		Path:        sproutsPath,
		VPath:       sproutsPath,
		BasePath:    sproutsPath,
		ShouldCrawl: shouldCrawl,
		ShouldMatch: shouldMatch,
		OnMatch:     onMatch,
	})
	if err != nil {
		return err
	}

	return nil
}

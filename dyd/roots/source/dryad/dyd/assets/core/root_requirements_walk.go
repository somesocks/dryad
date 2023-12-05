package core

import (
	fs2 "dryad/filesystem"
	"io/fs"
	"path/filepath"
)

func RootRequirementsWalk(path string, walkFn func(path string, info fs.FileInfo) error) error {
	path, err := RootPath(path)
	if err != nil {
		return err
	}

	requirementsPath := filepath.Join(path, "dyd", "requirements")

	requirementsExists, err := fileExists(requirementsPath)
	if err != nil {
		return err
	}

	// if requirements doesn't exist, do nothing
	if !requirementsExists {
		return nil
	}

	err = fs2.BFSWalk2(fs2.Walk4Request{
		Path:     requirementsPath,
		VPath:    requirementsPath,
		BasePath: requirementsPath,
		ShouldCrawl: func(context fs2.Walk4Context) (bool, error) {
			return context.Path == context.BasePath, nil
		},
		ShouldMatch: func(context fs2.Walk4Context) (bool, error) {
			return context.Path != context.BasePath, nil
		},
		OnMatch: func(context fs2.Walk4Context) error {
			return walkFn(context.Path, context.Info)
		},
	})
	if err != nil {
		return err
	}

	return nil
}

package core

import (
	fs2 "dryad/filesystem"
	"io/fs"
	"os"
	"path/filepath"
)

var _isRoot = func(path string, info fs.FileInfo) (bool, error) {

	typePath := filepath.Join(path, "dyd", "type")
	typeBytes, err := os.ReadFile(typePath)

	isRoot := err == nil && string(typeBytes) == "root"

	return isRoot, nil
}

func RootsWalk(path string, walkFn func(path string, info fs.FileInfo) error) error {
	var rootsPath, err = RootsPath(path)
	if err != nil {
		return err
	}

	err = fs2.Walk(fs2.WalkRequest{
		BasePath:     rootsPath,
		CrawlExclude: _isRoot,
		MatchInclude: _isRoot,
		OnMatch:      walkFn,
	})
	if err != nil {
		return err
	}

	return nil
}

package core

import (
	fs2 "dryad/filesystem"
	"io/fs"
	"os"
	"path/filepath"
)

var _isSprout = func(path string, info fs.FileInfo) (bool, error) {

	var dydPath = filepath.Join(path, "dyd", "fingerprint")
	var _, dydInfoErr = os.Stat(dydPath)
	var isSprout = dydInfoErr == nil

	return isSprout, nil
}

func SproutsWalk(path string, walkFn func(path string, info fs.FileInfo) error) error {
	var sproutsPath, err = SproutsPath(path)
	if err != nil {
		return err
	}

	err = fs2.Walk(fs2.WalkRequest{
		BasePath:     sproutsPath,
		CrawlExclude: _isSprout,
		MatchInclude: _isSprout,
		OnMatch:      walkFn,
	})
	if err != nil {
		return err
	}

	return nil
}

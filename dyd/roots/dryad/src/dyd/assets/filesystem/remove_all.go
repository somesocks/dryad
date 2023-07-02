package fs2

import (
	"io/fs"
	"os"
	"path/filepath"
)

func RemoveAll(path string) error {

	// walk through the filesystem and fix any permissions problems,
	// if you can
	err := DFSWalk(Walk3Request{
		BasePath: path,
		ShouldCrawl: func(path string, info fs.FileInfo) (bool, error) {
			// don't crawl symlinks
			if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				return false, nil
			}

			return true, nil
		},
		OnMatch: func(path string, info fs.FileInfo) error {
			parentInfo, err := os.Lstat(filepath.Dir(path))
			if err != nil {
				return err
			}

			if parentInfo.Mode()&0o200 != 0o200 {
				err := os.Chmod(filepath.Dir(path), parentInfo.Mode()|0o200)
				if err != nil {
					return err
				}
			}

			// if info.Mode()&0o200 != 0o200 {
			// 	err := os.Chmod(filepath.Dir(path), info.Mode()|0o200)
			// 	if err != nil {
			// 		return err
			// 	}
			// }

			err = os.Remove(path)
			if err != nil {
				return err
			}

			return nil
		},
	})
	if err != nil {
		return err
	}

	return nil
}

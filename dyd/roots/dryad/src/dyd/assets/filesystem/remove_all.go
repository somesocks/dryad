package fs2

import (
	"io/fs"
	"os"
)

func RemoveAll(path string) error {

	// walk through the filesystem and fix any permissions problems,
	// if you can
	err := Walk2(Walk2Request{
		BasePath: path,
		CrawlExclude: func(path string, info fs.FileInfo) (bool, error) {
			// don't crawl symlinks
			if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				return true, nil
			}

			return false, nil
		},
		OnMatch: func(path string, info fs.FileInfo) error {
			err := os.Chmod(path, os.ModePerm)
			if err != nil {
				return err
			}

			return nil
		},
	})
	if err != nil {
		return err
	}

	err = os.RemoveAll(path)
	if err != nil {
		return err
	}

	return nil
}

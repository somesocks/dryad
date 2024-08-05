package core

import (
	"io/fs"
	"os"
	"path/filepath"
)

func gardenRootsWalk(filename string, linkDirname string, walkFn filepath.WalkFunc) error {
	symWalkFunc := func(path string, info os.FileInfo, err error) error {

		if fname, err := filepath.Rel(filename, path); err == nil {
			path = filepath.Join(linkDirname, fname)
		} else {
			return err
		}

		if err == nil && info.Mode()&os.ModeSymlink == os.ModeSymlink {
			finalPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
			info, err := os.Lstat(finalPath)
			if err != nil {
				return walkFn(path, info, err)
			}
			if info.IsDir() {
				return gardenRootsWalk(finalPath, path, walkFn)
			}
		}

		return walkFn(path, info, err)
	}
	return filepath.Walk(filename, symWalkFunc)
}

func GardenRootsWalk(rootsPath string, walkFn filepath.WalkFunc) error {

	// log.Print("stem_path ", stem_path)
	err := gardenRootsWalk(
		rootsPath,
		rootsPath,
		func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				typePath := filepath.Join(path, "dyd", "type")
				_, fileInfoErr := os.Stat(typePath)

				if fileInfoErr == nil {
					err = walkFn(path, info, nil)
					if err != nil {
						return err
					} else {
						return filepath.SkipDir
					}
				}
			}

			return nil
		},
	)

	return err
}

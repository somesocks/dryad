package core

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func GardenPack(gardenPath string, targetPath string) (string, error) {
	var err error

	// convert relative stem path to absolute
	if !filepath.IsAbs(gardenPath) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		gardenPath = filepath.Join(wd, gardenPath)
	}

	// normalize garden path
	gardenPath, err = GardenPath(gardenPath)
	if err != nil {
		return "", err
	}

	// convert relative target to absolute
	if !filepath.IsAbs(targetPath) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		targetPath = filepath.Join(wd, targetPath)
	}

	// build archive name
	targetInfo, err := os.Stat(targetPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	} else if targetInfo.IsDir() {
		baseName := filepath.Base(gardenPath + ".tar.gz")
		targetPath = filepath.Join(targetPath, baseName)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var gzw = gzip.NewWriter(file)
	defer gzw.Close()

	var tw = tar.NewWriter(gzw)
	defer tw.Close()

	// var onMatch = func(walkPath string, info fs.FileInfo, pathErr error) error {
	// 	if pathErr != nil {
	// 		return pathErr
	// 	}

	// 	var relativePath string
	// 	var err error

	// 	relativePath, err = filepath.Rel(stemPath, walkPath)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// if we have a symlink, we need to read the real file to guarantee
	// 	// we get the real size and other info needed to build the tar header
	// 	if info.Mode()&os.ModeSymlink == os.ModeSymlink {
	// 		realPath, err := filepath.EvalSymlinks(walkPath)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		info, err = os.Stat(realPath)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}

	// 	// create a new dir/file header
	// 	header, err := tar.FileInfoHeader(info, relativePath)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	header.Name = relativePath

	// 	err = tw.WriteHeader(header)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if info.IsDir() {
	// 		return nil
	// 	}

	// 	file, err := os.Open(walkPath)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer file.Close()

	// 	_, err = io.Copy(tw, file)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	return nil
	// }

	// err = StemWalk(
	// 	StemWalkArgs{
	// 		BasePath: stemPath,
	// 		OnMatch:  onMatch,
	// 	},
	// )
	// if err != nil {
	// 	return "", err
	// }

	return targetPath, err
}

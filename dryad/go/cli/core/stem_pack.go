package core

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func StemPack(stemPath string, targetPath string) (string, error) {
	var err error

	// convert relative stem path to absolute
	if !filepath.IsAbs(stemPath) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		stemPath = filepath.Join(wd, stemPath)
	}

	// resolve the dir to the root of the stem
	stemPath, err = StemPath(stemPath)
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

	file, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var gzw = gzip.NewWriter(file)
	defer gzw.Close()

	var tw = tar.NewWriter(gzw)
	defer tw.Close()

	var onMatch = func(walkPath string, info fs.FileInfo, pathErr error) error {
		if pathErr != nil {
			return pathErr
		}

		var relativePath string
		var err error

		relativePath, err = filepath.Rel(stemPath, walkPath)
		if err != nil {
			return err
		}

		// if we have a symlink, we need to read the real file to guarantee
		// we get the real size and other info needed to build the tar header
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			realPath, err := filepath.EvalSymlinks(walkPath)
			if err != nil {
				return err
			}
			info, err = os.Stat(realPath)
			if err != nil {
				return err
			}
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(info, relativePath)
		if err != nil {
			return err
		}
		header.Name = relativePath

		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(walkPath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		if err != nil {
			return err
		}

		return nil
	}

	err = StemWalk(
		StemWalkArgs{
			BasePath: stemPath,
			OnMatch:  onMatch,
		},
	)
	if err != nil {
		return "", err
	}

	return targetPath, err
}

package core

import (
	"archive/tar"
	"compress/gzip"
	fs2 "dryad/filesystem"
	"errors"
	"fmt"
	"io"
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

	var packMap = make(map[string]bool)

	var packFile = func(path string, info fs.FileInfo) error {
		fmt.Println("packFile", path)
		var relativePath string
		var err error

		relativePath, err = filepath.Rel(gardenPath, path)
		if err != nil {
			return err
		}

		// don't pack a file that's already been packed
		if _, ok := packMap[relativePath]; ok {
			return nil
		}

		// if it's a symlink, run again on the real file
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			linkPath, err := os.Readlink(path)
			if err != nil {
				return err
			}

			// create a new dir/file header
			header, err := tar.FileInfoHeader(info, relativePath)
			if err != nil {
				return err
			}
			header.Name = relativePath
			header.Typeflag = tar.TypeSymlink
			header.Linkname = linkPath

			err = tw.WriteHeader(header)
			if err != nil {
				return err
			}

			// add path to the packMap
			packMap[relativePath] = true

		} else if info.IsDir() {
			// create a new dir/file header
			header, err := tar.FileInfoHeader(info, relativePath)
			if err != nil {
				return err
			}
			header.Name = relativePath
			header.Typeflag = tar.TypeDir

			err = tw.WriteHeader(header)
			if err != nil {
				return err
			}

			// add path to the packMap
			packMap[relativePath] = true
		} else if info.Mode().IsRegular() {
			// create a new dir/file header
			header, err := tar.FileInfoHeader(info, relativePath)
			if err != nil {
				return err
			}
			header.Name = relativePath
			header.Typeflag = tar.TypeReg

			err = tw.WriteHeader(header)
			if err != nil {
				return err
			}

			// add path to the packMap
			packMap[relativePath] = true

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tw, file)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = fs2.Walk2(
		fs2.Walk2Request{
			BasePath: filepath.Join(gardenPath),
			CrawlInclude: func(path string, info fs.FileInfo) (bool, error) {
				relPath, err := filepath.Rel(gardenPath, path)
				if err != nil {
					return false, err
				}

				return relPath == "." || relPath == "dyd", nil
			},
			MatchExclude: func(path string, info fs.FileInfo) (bool, error) {
				relPath, err := filepath.Rel(gardenPath, path)
				if err != nil {
					return false, err
				}

				if !info.IsDir() {
					return true, nil
				} else {
					return relPath == ".", nil
				}

			},
			OnMatch: packFile,
		},
	)
	if err != nil {
		return "", err
	}

	err = fs2.Walk2(
		fs2.Walk2Request{
			BasePath: filepath.Join(gardenPath, "dyd", "sprouts"),
			MatchExclude: func(path string, info fs.FileInfo) (bool, error) {
				relPath, err := filepath.Rel(gardenPath, path)
				if err != nil {
					return false, err
				}

				return relPath == "dyd/sprouts", nil
			},
			OnMatch: packFile,
		},
	)
	if err != nil {
		return "", err
	}

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

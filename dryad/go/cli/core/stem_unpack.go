package core

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func StemUnpack(gardenPath string, packPath string) (string, error) {
	var err error

	// clean garden path
	gardenPath, err = GardenPath(gardenPath)
	if err != nil {
		return "", err
	}

	// convert relative path to absolute
	if !filepath.IsAbs(packPath) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		packPath = filepath.Join(wd, packPath)
	}

	workspacePath, err := os.MkdirTemp("", "dryad-*")
	if err != nil {
		return "", err
	}
	// defer os.RemoveAll(workspacePath)
	fmt.Println("workspacePath", workspacePath)

	fr, err := os.Open(packPath)
	if err != nil {
		return "", err
	}

	gzr, err := gzip.NewReader(fr)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {

		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		path := filepath.Join(workspacePath, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return "", err
			}
		} else {
			if err = os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
				return "", err
			}

			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				return "", err
			}
			defer file.Close()
			_, err = io.Copy(file, tr)
			if err != nil {
				return "", err
			}
		}
	}

	stemFingerprint, err := StemValidate(workspacePath)
	if err != nil {
		return "", err
	}

	_, err = rootBuild_stage8(gardenPath, workspacePath, stemFingerprint)
	if err != nil {
		return stemFingerprint, err
	}

	return stemFingerprint, err
}

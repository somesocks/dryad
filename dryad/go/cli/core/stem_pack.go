package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func StemPack(stemPath string, targetPath string) (string, error) {
	var err error

	// resolve the dir to the root of the stem
	stemPath, err = StemPath(stemPath)
	if err != nil {
		return "", err
	}

	if targetPath == "" {
		targetPath, err = os.MkdirTemp("", "*")
	}
	if err != nil {
		return "", err
	}

	err = StemWalk(
		stemPath,
		func(walkPath string, info fs.FileInfo, pathErr error) error {
			fmt.Println("StemPack walk callback ", walkPath)
			if pathErr != nil {
				return pathErr
			}

			var relativePath string
			var err error

			relativePath, err = filepath.Rel(stemPath, walkPath)
			if err != nil {
				return err
			}

			var destPath = filepath.Join(targetPath, relativePath)

			if info.IsDir() {
				err = os.Mkdir(destPath, info.Mode().Perm())
				if err != nil {
					return err
				}
			} else {
				err = os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
				if err != nil {
					return err
				}

				var srcFile *os.File
				srcFile, err = os.Open(walkPath)
				if err != nil {
					return err
				}
				defer srcFile.Close()

				var destFile *os.File
				destFile, err = os.Create(destPath)
				if err != nil {
					return err
				}
				defer destFile.Close()

				fmt.Println("StemPack ", walkPath, " -> ", destPath)
				_, err = destFile.ReadFrom(srcFile)
				if err != nil {
					return err
				}

				err = destFile.Sync()
				if err != nil {
					return err
				}

				err = os.Chmod(destPath, info.Mode().Perm())
				if err != nil {
					return err
				}

			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	var fingerprintPath = filepath.Join(targetPath, "dyd", "fingerprint")
	var fingerprintExists bool

	fingerprintExists, err = fileExists(fingerprintPath)
	if err != nil {
		return "", err
	}

	if !fingerprintExists {
		var stemFingerprint string
		stemFingerprint, err = StemFingerprint(targetPath)
		if err != nil {
			return "", err
		}

		err = os.WriteFile(fingerprintPath, []byte(stemFingerprint), fs.ModePerm)
		if err != nil {
			return "", err
		}
	}

	return targetPath, err
}

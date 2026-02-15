package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func stemPathFromSprout(path string) (string, error) {
	dependencyPath := filepath.Join(path, "dyd", "dependencies", "stem")
	_, err := os.Lstat(dependencyPath)
	if err != nil {
		return "", err
	}

	resolvedDependencyPath, err := filepath.EvalSymlinks(dependencyPath)
	if err != nil {
		return "", err
	}

	return StemPath(resolvedDependencyPath)
}

func StemPath(path string) (string, error) {
	var working_path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}

	var dyd_path = filepath.Join(working_path, "dyd")
	var fileInfo, fileInfoErr = os.Stat(dyd_path)

	for working_path != "/" {

		if fileInfoErr == nil && fileInfo.IsDir() {
			typePath := filepath.Join(working_path, "dyd", "type")
			typeBytes, typeErr := os.ReadFile(typePath)
			if typeErr == nil && strings.TrimSpace(string(typeBytes)) == SentinelSprout.String() {
				return stemPathFromSprout(working_path)
			}

			return working_path, nil
		}

		working_path = filepath.Dir(working_path)
		dyd_path = filepath.Join(working_path, "dyd")
		fileInfo, fileInfoErr = os.Stat(dyd_path)
	}

	return "", errors.New("dyd stem path not found")
}

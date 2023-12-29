package core

import (
	"errors"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

func RootPath(path string) (string, error) {
	zlog.Trace().
		Str("path", path).
		Msg("RootPath")

	var err error
	path, err = filepath.Abs(path)
	// zlog.Trace().
	// 	Str("path", path).
	// 	Err(err).
	// 	Msg("RootPath.filepath.Abs")
	if err != nil {
		return "", err
	}

	var workingPath = path
	var flagPath = filepath.Join(workingPath, "dyd", "type")
	var fileBytes, fileInfoErr = os.ReadFile(flagPath)

	for workingPath != "/" {
		// zlog.Trace().
		// 	Str("workingPath", workingPath).
		// 	Str("flagPath", flagPath).
		// 	Err(fileInfoErr).
		// 	Msg("RootPath.workingPath")

		if fileInfoErr == nil && string(fileBytes) == "root" {

			zlog.Trace().
				Str("result", workingPath).
				Msg("RootPath success")

			return workingPath, nil
		}

		workingPath = filepath.Dir(workingPath)
		flagPath = filepath.Join(workingPath, "dyd", "type")
		fileBytes, fileInfoErr = os.ReadFile(flagPath)
	}

	zlog.Trace().
		Msg("RootPath failure")

	return "", errors.New("dyd root path not found starting from " + path)
}

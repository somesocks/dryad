package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

func RootPath(path string, limit string) (string, error) {
	zlog.Trace().
		Str("path", path).
		Msg("RootPath")

	var err error

	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}
	zlog.Trace().
		Str("path", path).
		Msg("RootPath/abs")

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	zlog.Trace().
		Str("path", path).
		Msg("RootPath/evalSym")

	var workingPath = path
	var flagPath = filepath.Join(workingPath, "dyd", "type")
	var fileBytes, fileInfoErr = os.ReadFile(flagPath)

	for workingPath != "/" && strings.HasPrefix(workingPath, limit) {
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

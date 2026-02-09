package core

import (
	"dryad/task"

	"errors"
	"os"
	"path/filepath"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

func _gardenPath(path string) (string, error) {
	zlog.Trace().
		Str("path", path).
		Msg("GardenPath")

	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}

	zlog.Trace().
		Str("path", path).
		Msg("GardenPath/abs")

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	zlog.Trace().
		Str("path", path).
		Msg("GardenPath/evalSym")

	var workingPath = path
	var flagPath = filepath.Join(workingPath, "dyd", "type")
	var fileBytes, fileInfoErr = os.ReadFile(flagPath)
	var sentinel Sentinel = SentinelFromString(strings.TrimSpace(string(fileBytes)))

	for workingPath != "/" {
		if fileInfoErr == nil {
			warnTypeFileWhitespace(filepath.Join("dyd", "type"), SentinelGarden.String(), string(fileBytes))
		}

		if fileInfoErr == nil && sentinel == SentinelGarden {
			return workingPath, nil
		}

		workingPath = filepath.Dir(workingPath)
		flagPath = filepath.Join(workingPath, "dyd", "type")
		fileBytes, fileInfoErr = os.ReadFile(flagPath)
		sentinel = SentinelFromString(strings.TrimSpace(string(fileBytes)))
	}

	return "", errors.New("dyd garden path not found starting from " + path)
}

func (ug *UnsafeGardenReference) Resolve(ctx *task.ExecutionContext) (error, *SafeGardenReference) {
	zlog.Trace().
		Str("BasePath", ug.BasePath).
		Msg("UnsafeGardenReference.Resolve")

	var gardenPath string = ug.BasePath
	var safeRef SafeGardenReference
	var err error

	gardenPath, err = _gardenPath(ug.BasePath)
	if err != nil {
		return err, nil
	}

	safeRef = SafeGardenReference{BasePath: gardenPath}

	return nil, &safeRef
}

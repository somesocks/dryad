
package core

import (
	"dryad/task"

	"errors"
	"os"
	"path/filepath"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

func rootPath(path string, limit string) (string, error) {
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

func (ur *UnsafeRootReference) Resolve(ctx * task.ExecutionContext, _ any) (error, SafeRootReference) {
	var gardenPath string = ur.Roots.Garden.BasePath
	var basePath string = ur.BasePath
	var err error

	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(gardenPath, basePath)
	}

	// resolve the path to the base of the root
	basePath, err = rootPath(basePath, gardenPath) 
	if err != nil {
		return err, SafeRootReference{}
	}
	return nil, SafeRootReference{
		BasePath: basePath,
		Roots: ur.Roots,
	}
}

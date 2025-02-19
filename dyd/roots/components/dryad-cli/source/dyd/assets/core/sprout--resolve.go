
package core

import (
	"os"
	"dryad/task"
	"path/filepath"

	"strings"
	"errors"

	// zlog "github.com/rs/zerolog/log"

)

func sproutPath(ctx *task.ExecutionContext, path string, limit string) (error, string) {
	var workingPath string = path

	var dydPath = filepath.Join(workingPath, "dyd")
	var fileInfo, fileInfoErr = os.Stat(dydPath)

	for workingPath != "/" && strings.HasPrefix(workingPath, limit) {

		if fileInfoErr == nil && fileInfo.IsDir() {
			return nil, workingPath
		}

		workingPath = filepath.Dir(workingPath)
		dydPath = filepath.Join(workingPath, "dyd")
		fileInfo, fileInfoErr = os.Stat(dydPath)
	}

	return errors.New("dyd sprout path not found"), ""
}


func (sprout *UnsafeSproutReference) Resolve(ctx * task.ExecutionContext) (error, *SafeSproutReference) {
	var basePath string = sprout.BasePath
	var err error
	var res SafeSproutReference

	// resolve the path to the base of the sprout, within the sprouts
	err, basePath = sproutPath(ctx, basePath, sprout.Sprouts.BasePath) 
	if err != nil {
		return err, nil
	}

	res = SafeSproutReference{
		BasePath: basePath,
		Sprouts: sprout.Sprouts,
	}

	return nil, &res
}

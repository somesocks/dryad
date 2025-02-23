
package core

import (
	"dryad/task"
	"path/filepath"

	// zlog "github.com/rs/zerolog/log"

)

func (ur *UnsafeRootReference) Resolve(ctx * task.ExecutionContext, _ any) (error, SafeRootReference) {
	var gardenPath string = ur.Roots.Garden.BasePath
	var basePath string = ur.BasePath
	var err error

	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(gardenPath, basePath)
	}

	// resolve the path to the base of the root
	basePath, err = RootPath(basePath, gardenPath) 
	if err != nil {
		return err, SafeRootReference{}
	}
	return nil, SafeRootReference{
		BasePath: basePath,
		Roots: ur.Roots,
	}
}

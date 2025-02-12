
package core

import (
	"dryad/task"
	"path/filepath"
)

func (ur *UnsafeRootReference) Resolve(ctx * task.ExecutionContext, _ any) (error, SafeRootReference) {
	var gardenPath string = ur.Garden.BasePath
	var basePath string = ur.BasePath
	var err error

	// convert filepath to absolute (relative to the garden)
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
		Garden: ur.Garden,
	}
}

func (ur *UnsafeRootReference) Clean() (UnsafeRootReference) {
	var gardenPath string = ur.Garden.BasePath
	var basePath string = ur.BasePath

	// convert filepath to absolute (relative to the garden)
	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(gardenPath, basePath)
	}

	return UnsafeRootReference{
		BasePath: basePath,
		Garden: ur.Garden,
	}
}
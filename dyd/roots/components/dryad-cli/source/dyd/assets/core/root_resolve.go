
package core

import (
	"os"
	"dryad/task"
	"path/filepath"

	// zlog "github.com/rs/zerolog/log"

)

func (ur *UnsafeRootReference) Resolve(ctx * task.ExecutionContext, _ any) (error, SafeRootReference) {
	var gardenPath string = ur.Roots.Garden.BasePath
	var basePath string = ur.BasePath
	var err error

	// convert root base path to absolute
	if !filepath.IsAbs(basePath) {
		wd, err := os.Getwd()
		if err != nil {
			return err, SafeRootReference{}
		}

		// the base of the path needs to cleaned of symlinks, 
		// to make sure it matches the garden path
		wd, err = filepath.EvalSymlinks(wd)
		if err != nil {
			return err, SafeRootReference{}
		}

		basePath = filepath.Join(wd, basePath)
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

func (ur *UnsafeRootReference) Clean() (error, UnsafeRootReference) {
	var basePath string = ur.BasePath

	// convert root base path to absolute
	if !filepath.IsAbs(basePath) {
		wd, err := os.Getwd()
		if err != nil {
			return err, UnsafeRootReference{}
		}

		// the base of the path needs to cleaned of symlinks, 
		// to make sure it matches the garden path
		wd, err = filepath.EvalSymlinks(wd)
		if err != nil {
			return err, UnsafeRootReference{}
		}

		basePath = filepath.Join(wd, basePath)
	}

	return nil, UnsafeRootReference{
		BasePath: basePath,
		Roots: ur.Roots,
	}
}
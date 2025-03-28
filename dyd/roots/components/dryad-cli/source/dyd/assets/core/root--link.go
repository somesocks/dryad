package core

import (
	"os"
	"path/filepath"
	"dryad/task"
)


type RootLinkRequest struct {
	Dependency *SafeRootReference
	Alias string
}

func (root *SafeRootReference) Link(ctx *task.ExecutionContext, req RootLinkRequest) (error) {
	var alias string = req.Alias
	var depBasePath string = req.Dependency.BasePath
	var err error

	if alias == "" {
		alias = filepath.Base(depBasePath)
	}

	var requirementsPath = filepath.Join(root.BasePath, "dyd", "requirements")
	var aliasPath = filepath.Join(requirementsPath, alias)

	// make sure the roots path exists before trying to link
	err = os.MkdirAll(requirementsPath, os.ModePerm)
	if err != nil {
		return err
	}

	var linkPath string
	linkPath, err = filepath.Rel(requirementsPath, depBasePath)
	if err != nil {
		return err
	}

	err = os.Symlink(linkPath, aliasPath)
	if err != nil {
		return err
	}

	return nil
}
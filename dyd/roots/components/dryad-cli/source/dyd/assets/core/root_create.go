package core

import (
	"fmt"
	"os"
	"path/filepath"

	"dryad/task"
)

type RootCreateRequest struct {
	Root *UnsafeRootReference
}


func RootCreate(ctx *task.ExecutionContext, req RootCreateRequest) (error, *SafeRootReference) {
	var path string = req.Root.BasePath

	// check to make sure the destination doesn't already exist
	pathExists, err := fileExists(path)
	if err != nil {
		return err, nil
	} else if pathExists {
		return fmt.Errorf("error: root destination %s already exists", path), nil
	}

	// check to make sure that the destination is within roots dir
	rootsPath := req.Root.Roots.BasePath

	isInRootsDir, err := fileIsDescendant(path, rootsPath)
	if err != nil {
		return err, nil
	} else if !isInRootsDir {
		return fmt.Errorf("error: root destination %s must be in roots directory %s", path, rootsPath), nil
	}

	var basePath string = filepath.Join(path, "dyd")
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		return err, nil
	}

	// write out type file
	typePath := filepath.Join(basePath, "type")

	typeFile, err := os.Create(typePath)
	if err != nil {
		return err, nil
	}
	defer typeFile.Close()

	_, err = fmt.Fprint(typeFile, "root")
	if err != nil {
		return err, nil
	}

	var assetsPath string = filepath.Join(basePath, "assets")
	if err := os.MkdirAll(assetsPath, os.ModePerm); err != nil {
		return err, nil
	}

	var commandsPath string = filepath.Join(basePath, "commands")
	if err := os.MkdirAll(commandsPath, os.ModePerm); err != nil {
		return err, nil
	}

	var docsPath string = filepath.Join(basePath, "docs")
	if err := os.MkdirAll(docsPath, os.ModePerm); err != nil {
		return err, nil
	}

	var requirementsPath string = filepath.Join(basePath, "requirements")
	if err := os.MkdirAll(requirementsPath, os.ModePerm); err != nil {
		return err, nil
	}

	var traitsPath string = filepath.Join(basePath, "traits")
	if err := os.MkdirAll(traitsPath, os.ModePerm); err != nil {
		return err, nil
	}

	var secretsPath string = filepath.Join(basePath, "secrets")
	if err := os.MkdirAll(secretsPath, os.ModePerm); err != nil {
		return err, nil
	}

	var defaultCommandPath string = filepath.Join(basePath, "commands", "default")
	if _, err := os.Create(defaultCommandPath); err != nil {
		return err, nil
	}

	if err := os.Chmod(defaultCommandPath, 0775); err != nil {
		return err, nil
	}

	var safeRoot SafeRootReference

	err, safeRoot = req.Root.Resolve(ctx, nil)
	if err != nil {
		return err, nil
	}

	return nil, &safeRoot
}

package core

import (
	dydfs "dryad/filesystem"
	"dryad/internal/filepath"
	"dryad/task"

	"dryad/internal/os"
	"io/fs"
)

func rootBuild_pathPopulate(workspacePath string, pathPath string) error {
	// walk through the dependencies, build them, and add the fingerprint as a dependency
	dependenciesPath := filepath.Join(workspacePath, "dyd", "dependencies")

	dependencies, err := filepath.Glob(filepath.Join(dependenciesPath, "*"))
	if err != nil {
		return err
	}

	for _, dependencyPath := range dependencies {
		baseName := filepath.Base(dependencyPath)

		commandsPath := filepath.Join(dependencyPath, "dyd", "commands")
		commands, err := filepath.Glob(filepath.Join(commandsPath, "*"))
		if err != nil {
			return err
		}

		for _, commandPath := range commands {
			commandName := filepath.Base(commandPath)
			baseTemplate := rootBuild_pathStub(baseName, commandName)

			var stubName string
			if commandName == "dyd-stem-run" {
				stubName = baseName
			} else if commandName == "default" {
				stubName = baseName
			} else {
				stubName = baseName + "--" + commandName
			}

			err = os.WriteFile(
				filepath.Join(pathPath, stubName),
				[]byte(baseTemplate),
				fs.ModePerm,
			)
			if err != nil {
				return err
			}

		}

	}

	return nil
}

func rootBuild_pathPrepare(workspacePath string) error {
	pathPath := filepath.Join(workspacePath, "dyd", "path")

	err, _ := dydfs.RemoveAll(task.SERIAL_CONTEXT, pathPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(pathPath, fs.ModePerm)
	if err != nil {
		return err
	}

	return rootBuild_pathPopulate(workspacePath, pathPath)
}

func rootBuild_pathPrepareFresh(workspacePath string) error {
	pathPath := filepath.Join(workspacePath, "dyd", "path")

	if err := os.Mkdir(pathPath, fs.ModePerm); err != nil {
		return err
	}

	return rootBuild_pathPopulate(workspacePath, pathPath)
}

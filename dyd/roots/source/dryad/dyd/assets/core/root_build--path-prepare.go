package core

import (
	dydfs "dryad/filesystem"

	"io/fs"
	"os"
	"path/filepath"
)

func rootBuild_pathPrepare(workspacePath string) error {

	pathPath := filepath.Join(workspacePath, "dyd", "path")

	err := dydfs.RemoveAll(pathPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(pathPath, fs.ModePerm)
	if err != nil {
		return err
	}

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
			if commandName == "default" {
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

package core

import (
	"dryad/internal/filepath"
	"dryad/internal/os"
)

type rootBuild_linkDependencyRequest struct {
	WorkspacePath         string
	DependencyName        string
	DependencyHeapPath    string
	DependencyFingerprint string
}

func rootBuild_linkDependency(req rootBuild_linkDependencyRequest) error {
	dependenciesPath := filepath.Join(req.WorkspacePath, "dyd", "dependencies")
	requirementsPath := filepath.Join(req.WorkspacePath, "dyd", "requirements")
	pathPath := filepath.Join(req.WorkspacePath, "dyd", "path")

	dependencyPath := filepath.Join(dependenciesPath, req.DependencyName)
	if err := os.Symlink(req.DependencyHeapPath, dependencyPath); err != nil {
		return err
	}

	if err := os.WriteFile(
		filepath.Join(requirementsPath, req.DependencyName),
		[]byte(req.DependencyFingerprint),
		0o511,
	); err != nil {
		return err
	}

	commandsPath := filepath.Join(req.DependencyHeapPath, "dyd", "commands")
	commands, err := os.ReadDir(commandsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, command := range commands {
		commandName := command.Name()
		stubName := req.DependencyName + "--" + commandName
		if commandName == "dyd-stem-run" || commandName == "default" {
			stubName = req.DependencyName
		}

		if err := os.WriteFile(
			filepath.Join(pathPath, stubName),
			[]byte(rootBuild_pathStub(req.DependencyName, commandName)),
			os.ModePerm,
		); err != nil {
			return err
		}
	}

	return nil
}

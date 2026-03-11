package core

import (
	dydfs "dryad/filesystem"
	"dryad/internal/filepath"
	"dryad/task"

	"dryad/internal/os"
	"io"
)

func sproutRequirementsCopyFile(sourcePath string, destPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return err
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Chmod(0o511)
}

func sproutRequirementsPrepare(sproutPath string) error {
	requirementsPath := filepath.Join(sproutPath, "dyd", "requirements")

	err, _ := dydfs.RemoveAll(task.SERIAL_CONTEXT, requirementsPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(requirementsPath, os.ModePerm); err != nil {
		return err
	}

	dependenciesPath := filepath.Join(sproutPath, "dyd", "dependencies")
	dependencyEntries, err := os.ReadDir(dependenciesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, dependencyEntry := range dependencyEntries {
		dependencyName := dependencyEntry.Name()
		dependencySourcePath := filepath.Join(dependenciesPath, dependencyName)

		resolvedDependencyPath, err := filepath.EvalSymlinks(dependencySourcePath)
		if err != nil {
			return err
		}

		stemDependencyPath, err := StemPath(resolvedDependencyPath)
		if err != nil {
			return err
		}

		if err := sproutRequirementsCopyFile(
			filepath.Join(stemDependencyPath, "dyd", "fingerprint"),
			filepath.Join(requirementsPath, dependencyName),
		); err != nil {
			return err
		}
	}

	return nil
}

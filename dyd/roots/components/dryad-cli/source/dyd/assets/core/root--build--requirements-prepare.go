package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"
	"io"

	"dryad/internal/filepath"
	"dryad/internal/os"
)

var rootBuild_requirementsPopulate = func() func(string, string) error {
	copyFingerprint := func(sourcePath string, destPath string) error {
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

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

	return func(workspacePath string, requirementsPath string) error {
		dependenciesPath := filepath.Join(workspacePath, "dyd", "dependencies")
		dependencyEntries, err := os.ReadDir(dependenciesPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		for _, dependencyEntry := range dependencyEntries {
			dependencyName := dependencyEntry.Name()
			dependencyPath := filepath.Join(dependenciesPath, dependencyName)

			err = copyFingerprint(
				filepath.Join(dependencyPath, "dyd", "fingerprint"),
				filepath.Join(requirementsPath, dependencyName),
			)
			if err != nil {
				return err
			}
		}

		return nil
	}

}()

// This function is used to prepare the dyd/requirements for a package,
// based on the contents of dyd/dependencies.
var rootBuild_requirementsPrepare = func(workspacePath string) error {
	requirementsPath := filepath.Join(workspacePath, "dyd", "requirements")

	err, _ := dydfs.RemoveAll(task.SERIAL_CONTEXT, requirementsPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(requirementsPath, os.ModePerm); err != nil {
		return err
	}

	return rootBuild_requirementsPopulate(workspacePath, requirementsPath)
}

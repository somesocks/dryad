package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"
	"io"

	"dryad/internal/filepath"
	"dryad/internal/os"
)

// This function is used to prepare the dyd/requirements for a package,
// based on the contents of dyd/dependencies.
var rootBuild_requirementsPrepare = func() func(string) error {
	copyFingerprint := func(sourcePath string, destPath string) error {
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

	var action = func(workspacePath string) error {
		requirementsPath := filepath.Join(workspacePath, "dyd", "requirements")

		err, _ := dydfs.RemoveAll(task.SERIAL_CONTEXT, requirementsPath)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(requirementsPath, os.ModePerm); err != nil {
			return err
		}

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

			resolvedDependencyPath, err := filepath.EvalSymlinks(dependencyPath)
			if err != nil {
				return err
			}

			stemDependencyPath, err := StemPath(resolvedDependencyPath)
			if err != nil {
				return err
			}

			err = copyFingerprint(
				filepath.Join(stemDependencyPath, "dyd", "fingerprint"),
				filepath.Join(requirementsPath, dependencyName),
			)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return action

}()

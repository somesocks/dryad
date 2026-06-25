package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"
	"fmt"
	"io"

	"dryad/internal/filepath"
	"dryad/internal/os"
)

var rootBuild_requirementsPopulate = func() func(string, string) error {
	copyFingerprint := func(sourcePath string, destPath string) error {
		destExists, err := fileExists(destPath)
		if err != nil {
			return err
		}
		if destExists {
			return fmt.Errorf("duplicate materialized requirement name: %s", filepath.Base(destPath))
		}

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

func rootBuild_collectEnvRequirements(requirementsPath string) (map[string][]byte, error) {
	envRequirements := map[string][]byte{}
	requirementEntries, err := os.ReadDir(requirementsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return envRequirements, nil
		}
		return nil, err
	}

	for _, requirementEntry := range requirementEntries {
		if requirementEntry.IsDir() {
			continue
		}

		requirementName := requirementEntry.Name()
		requirementPath := filepath.Join(requirementsPath, requirementName)
		requirementBytes, err := os.ReadFile(requirementPath)
		if err != nil {
			return nil, err
		}

		err, _, isEnv := rootRequirementParseEnvTarget(string(requirementBytes))
		if err != nil {
			return nil, fmt.Errorf("invalid env requirement %s: %w", requirementName, err)
		}
		if !isEnv {
			continue
		}

		err, alias, condition := rootRequirementParseName(requirementName)
		if err != nil {
			return nil, err
		}
		if len(condition) > 0 {
			return nil, fmt.Errorf("conditional env requirement is not valid in a stem: %s", requirementName)
		}

		err, injectName := rootRequirementCanonicalEnvName(alias)
		if err != nil {
			return nil, err
		}
		if _, exists := envRequirements[injectName]; exists {
			return nil, fmt.Errorf("duplicate materialized env requirement name: %s", injectName)
		}

		envRequirements[injectName] = requirementBytes
	}

	return envRequirements, nil
}

// This function is used to prepare the dyd/requirements for a package,
// based on the contents of dyd/dependencies.
var rootBuild_requirementsPrepare = func(workspacePath string) error {
	requirementsPath := filepath.Join(workspacePath, "dyd", "requirements")
	envRequirements, err := rootBuild_collectEnvRequirements(requirementsPath)
	if err != nil {
		return err
	}

	err, _ = dydfs.RemoveAll(task.SERIAL_CONTEXT, requirementsPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(requirementsPath, os.ModePerm); err != nil {
		return err
	}

	for requirementName, requirementBytes := range envRequirements {
		if err := os.WriteFile(filepath.Join(requirementsPath, requirementName), requirementBytes, 0o511); err != nil {
			return err
		}
	}

	return rootBuild_requirementsPopulate(workspacePath, requirementsPath)
}

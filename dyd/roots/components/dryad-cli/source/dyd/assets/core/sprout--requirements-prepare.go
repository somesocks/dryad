package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"

	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

func sproutRequirementsCopyTree(sourcePath string, destPath string, dependencyPath string) error {
	return filepath.WalkDir(sourcePath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		destEntryPath := filepath.Join(destPath, relPath)
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return os.MkdirAll(destEntryPath, os.ModePerm)
		}

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}

			absLinkTarget := linkTarget
			if !filepath.IsAbs(absLinkTarget) {
				absLinkTarget = filepath.Clean(filepath.Join(filepath.Dir(path), absLinkTarget))
			}

			isInternalLink, err := fileIsDescendant(absLinkTarget, dependencyPath)
			if err != nil {
				return err
			}
			if !isInternalLink {
				return errors.New("sprout requirements prepare - dependency symlink escapes dependency root")
			}

			if err := os.MkdirAll(filepath.Dir(destEntryPath), os.ModePerm); err != nil {
				return err
			}

			return os.Symlink(linkTarget, destEntryPath)
		}

		if info.Mode().IsRegular() {
			return sproutRequirementsCopyFile(path, destEntryPath)
		}

		return errors.New("sprout requirements prepare - unsupported file type in dependency traits")
	})
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

		requirementDependencyPath := filepath.Join(requirementsPath, dependencyName, "dyd")
		if err := os.MkdirAll(requirementDependencyPath, os.ModePerm); err != nil {
			return err
		}

		if err := sproutRequirementsCopyFile(
			filepath.Join(stemDependencyPath, "dyd", "fingerprint"),
			filepath.Join(requirementDependencyPath, "fingerprint"),
		); err != nil {
			return err
		}

		dependencyTraitsPath := filepath.Join(stemDependencyPath, "dyd", "traits")
		dependencyTraitsExists, err := fileExists(dependencyTraitsPath)
		if err != nil {
			return err
		}
		if !dependencyTraitsExists {
			continue
		}

		if err := sproutRequirementsCopyTree(
			dependencyTraitsPath,
			filepath.Join(requirementDependencyPath, "traits"),
			stemDependencyPath,
		); err != nil {
			return err
		}
	}

	return nil
}

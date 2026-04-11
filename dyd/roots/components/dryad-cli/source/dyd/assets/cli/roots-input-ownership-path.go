package cli

import (
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/internal/filepath"
	"dryad/task"
	"strings"
)

var rootsInputOwnershipDependencyCorrection = func(path string) string {
	p1, _ := filepath.Split(path)
	p1 = filepath.Clean(p1)
	p2, f2 := filepath.Split(p1)
	p2 = filepath.Clean(p2)
	p3, f3 := filepath.Split(p2)
	p3 = filepath.Clean(p3)

	if f3 == "dyd" && (f2 == "requirements" || strings.HasPrefix(f2, "requirements"+dryad.RootRequirementSelectorSeparator)) {
		return p3
	}
	return path
}

func rootsInputOwnershipPaths(ctx *task.ExecutionContext, rawPath string) (error, string, string) {
	path, err := filepath.Abs(rawPath)
	if err != nil {
		return err, "", ""
	}

	correctedPath := rootsInputOwnershipDependencyCorrection(path)
	err, owningPath := dydfs.PartialEvalSymlinks(ctx, correctedPath)
	if err != nil {
		return err, "", ""
	}

	changedPath := owningPath
	if correctedPath != path {
		relPath, err := filepath.Rel(correctedPath, path)
		if err != nil {
			return err, "", ""
		}
		changedPath = filepath.Join(owningPath, relPath)
	}

	return nil, owningPath, changedPath
}

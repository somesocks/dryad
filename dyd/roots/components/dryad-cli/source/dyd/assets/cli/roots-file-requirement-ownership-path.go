package cli

import (
	"context"
	dydfs "dryad/filesystem"
	"dryad/internal/filepath"
)

func rootsFileRequirementOwnershipPath(ctx context.Context, rawPath string) (error, string) {
	path, err := filepath.Abs(rawPath)
	if err != nil {
		return err, ""
	}

	parentPath := filepath.Dir(path)
	err, parentPath = dydfs.PartialEvalSymlinks(ctx, parentPath)
	if err != nil {
		return err, ""
	}

	return nil, filepath.Join(parentPath, filepath.Base(path))
}

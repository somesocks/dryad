package core

import (
	// dydfs "dryad/filesystem"

	// "io/fs"
	// "io/ioutil"
	// "os"
	"context"
	"dryad/internal/filepath"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage3_request struct {
	RootPath      string
	WorkspacePath string
	GardenPath    string
}

// stage 3 - finalize the stem by generating fingerprints,
var rootBuild_stage3 func(ctx context.Context, req rootBuild_stage3_request) (error, string) = func(ctx context.Context, req rootBuild_stage3_request) (error, string) {
	relRootPath, err := filepath.Rel(
		filepath.Join(req.GardenPath, "dyd", "roots"),
		req.RootPath,
	)
	if err != nil {
		return err, ""
	}

	zlog.Trace().
		Str("path", relRootPath).
		Msg("root build - stage3")

	err, stemFingerprint := stemFinalize(ctx, req.WorkspacePath)
	return err, stemFingerprint
}

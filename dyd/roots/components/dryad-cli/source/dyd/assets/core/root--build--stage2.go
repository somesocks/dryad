package core

import (
	"dryad/task"

	// "io/fs"
	// "io/ioutil"
	"dryad/internal/filepath"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage2_request struct {
	RootPath   string
	GardenPath string
}

// stage 2 - generate the artificial links to all executable stems for the path
var rootBuild_stage2 func(ctx *task.ExecutionContext, req rootBuild_stage2_request) (error, any)

func init() {

	rootBuild_stage2 = func(ctx *task.ExecutionContext, req rootBuild_stage2_request) (error, any) {
		relRootPath, err := filepath.Rel(
			filepath.Join(req.GardenPath, "dyd", "roots"),
			req.RootPath,
		)
		if err != nil {
			return err, nil
		}
		zlog.Trace().
			Str("path", relRootPath).
			Msg("RootBuild/stage2")

		return nil, nil
	}

}

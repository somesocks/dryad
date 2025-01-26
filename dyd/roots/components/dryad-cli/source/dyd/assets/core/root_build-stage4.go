package core

import (
	// dydfs "dryad/filesystem"
	"dryad/task"

	// "io/fs"
	// "io/ioutil"
	// "os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage4_request struct {
	RootPath string
	WorkspacePath string
	GardenPath string
}

// stage 4 - check the garden to see if the stem exists,
// and add it if it doesn't
var rootBuild_stage4 func (ctx *task.ExecutionContext, req rootBuild_stage4_request) (error, string) =
	func (ctx *task.ExecutionContext, req rootBuild_stage4_request) (error, string) {
		relRootPath, err := filepath.Rel(
			filepath.Join(req.GardenPath, "dyd", "roots"),
			req.RootPath,
		)
		if err != nil {
			return err, ""
		}

		zlog.Debug().
			Str("path", relRootPath).
			Msg("root build - stage4")

		err, stemPath := HeapAddStem(
			ctx,
			HeapAddStemRequest{
				HeapPath: req.GardenPath,
				StemPath: req.WorkspacePath,
			},
		)
		return err, stemPath
	}

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

type rootBuild_stage3_request struct {
	Context *BuildContext
	RootPath string
	WorkspacePath string
	GardenPath string
}

// stage 3 - finalize the stem by generating fingerprints,
var rootBuild_stage3 func (ctx *task.ExecutionContext, req rootBuild_stage3_request) (error, string) =
	func (ctx *task.ExecutionContext, req rootBuild_stage3_request) (error, string) {
		relRootPath, err := filepath.Rel(
			filepath.Join(req.GardenPath, "dyd", "roots"),
			req.RootPath,
		)
		if err != nil {
			return err, ""
		}

		zlog.Debug().
			Str("path", relRootPath).
			Msg("root build - stage3")

		stemFingerprint, err := stemFinalize(req.WorkspacePath)
		return err, stemFingerprint
	}

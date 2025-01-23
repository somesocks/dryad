package core

import (
	// dydfs "dryad/filesystem"
	"dryad/task"

	// "io/fs"
	// "io/ioutil"
	// "os"
	// "path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage6_request struct {
	RelRootPath string
	StemBuildPath string
	GardenPath string
}

// stage 6 - pack the derived stem into the heap and garden
var rootBuild_stage6 func (ctx *task.ExecutionContext, req rootBuild_stage6_request) (error, string) =
	func (ctx *task.ExecutionContext, req rootBuild_stage6_request) (error, string) {
		zlog.Debug().
			Str("path", req.RelRootPath).
			Msg("root build - stage6")

		err, stemPath := HeapAddStem(
			ctx,
			HeapAddStemRequest{
				HeapPath: req.GardenPath,
				StemPath: req.StemBuildPath,
			},
		)
		return err, stemPath
	}

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
	Garden *SafeGardenReference
	RootPath string
	WorkspacePath string
}

// stage 4 - check the garden to see if the stem exists,
// and add it if it doesn't
var rootBuild_stage4 func (ctx *task.ExecutionContext, req rootBuild_stage4_request) (error, *SafeHeapStemReference) =
	func (ctx *task.ExecutionContext, req rootBuild_stage4_request) (error, *SafeHeapStemReference) {
		relRootPath, err := filepath.Rel(
			filepath.Join(req.Garden.BasePath, "dyd", "roots"),
			req.RootPath,
		)
		if err != nil {
			return err, nil
		}

		err, heap := req.Garden.Heap().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, stems := heap.Stems().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		zlog.Trace().
			Str("path", relRootPath).
			Msg("root build - stage4")

		err = sanitizeTypeFile(
			filepath.Join(req.RootPath, "dyd", "type"),
			SentinelRoot.String(),
		)
		if err != nil {
			return err, nil
		}

		err, stem := stems.AddStem(
			ctx,
			HeapAddStemRequest{
				StemPath: req.WorkspacePath,
			},
		)
		return err, stem
	}

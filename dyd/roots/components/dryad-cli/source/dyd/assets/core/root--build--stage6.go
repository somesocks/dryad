package core

import (

	// "io/fs"
	// "io/ioutil"
	// "os"
	// "dryad/internal/filepath"

	"context"

	zlog "github.com/rs/zerolog/log"
)

type rootBuild_stage6_request struct {
	Garden        *SafeGardenReference
	RelRootPath   string
	StemBuildPath string
}

// stage 6 - pack the derived stem into the heap and garden
var rootBuild_stage6 func(ctx context.Context, req rootBuild_stage6_request) (error, *SafeHeapStemReference) = func(ctx context.Context, req rootBuild_stage6_request) (error, *SafeHeapStemReference) {
	zlog.Trace().
		Str("path", req.RelRootPath).
		Msg("root build - stage6")

	err, heap := req.Garden.Heap().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	err, stems := heap.Stems().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	err, stem := stems.AddStem(
		ctx,
		HeapAddStemRequest{
			StemPath: req.StemBuildPath,
		},
	)
	return err, stem
}

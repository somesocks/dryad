package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	"errors"
	"io/fs"

	"dryad/internal/os"
	"path/filepath"
	// zlog "github.com/rs/zerolog/log"
)

func (heapDerivations *UnsafeHeapDerivationsReference) Resolve(ctx *task.ExecutionContext) (error, *SafeHeapDerivationsReference) {
	var heapDerivationsExists bool
	var err error
	var safeRef SafeHeapDerivationsReference

	heapDerivationsExists, err = fileExists(heapDerivations.BasePath)
	if err != nil {
		return err, nil
	}

	if !heapDerivationsExists {
		err, _ := fs2.Mkdir2(
			ctx,
			fs2.MkdirRequest{
				Path: heapDerivations.BasePath,
				Mode: os.ModePerm,
			},
		)
		if err != nil {
			return err, nil
		}
	}

	rootsPath := filepath.Join(heapDerivations.BasePath, "roots")
	rootsInfo, err := os.Lstat(rootsPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err, nil
		}

		err, _ = fs2.Mkdir2(
			ctx,
			fs2.MkdirRequest{
				Path:      rootsPath,
				Mode:      os.ModePerm,
				Recursive: true,
			},
		)
		if err != nil {
			return err, nil
		}
	} else if !rootsInfo.IsDir() {
		return fs.ErrInvalid, nil
	}

	safeRef = SafeHeapDerivationsReference{
		BasePath: heapDerivations.BasePath,
		Heap:     heapDerivations.Heap,
	}

	return nil, &safeRef
}

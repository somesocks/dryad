package core

import (
	"dryad/task"
	"errors"
	"io/fs"
	"os"
	// "errors"
	// "path/filepath"
	// "os"
	// zlog "github.com/rs/zerolog/log"
)

func (heapDerivation *UnsafeHeapDerivationReference) Exists(ctx *task.ExecutionContext) (error, bool) {
	info, err := os.Lstat(heapDerivation.BasePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false
		}
		return err, false
	}

	return nil, info.Mode().IsRegular()
}

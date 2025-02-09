package core

import (
	"os"
	"errors"
	"io/fs"
	"path/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

type HeapAddFileRequest struct {
	Garden *SafeGardenReference
	SourcePath string
}

func HeapAddFile(ctx *task.ExecutionContext, req HeapAddFileRequest) (error, string) {
	var heapPath string
	var sourcePath string = req.SourcePath
	var err error

	heapPath, err = HeapPath(req.Garden)
	if err != nil {
		return err, ""
	}

	zlog.
		Trace().
		Str("heapPath", heapPath).
		Str("sourcePath", sourcePath).
		Msg("HeapAddFile")

	sourceHashAlgorithm, sourceHash, err := fileHash(sourcePath)
	if err != nil {
		return err, ""
	}

	fingerprint := sourceHashAlgorithm + "-" + sourceHash

	destPath := filepath.Join(heapPath, "files", fingerprint)

	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return err, ""
	}
	defer srcFile.Close()

	var destFile *os.File
	destFile, err = os.OpenFile(destPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		// return a success if the file is already in the heap
		if errors.Is(err, fs.ErrExist) {
			return nil, fingerprint
		} else {
			return err, ""
		}
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(srcFile)
	if err != nil {
		return err, ""
	}

	// heap files should be set to R-X--X--X
	err = destFile.Chmod(0o511)
	if err != nil {
		return err, ""
	}

	return nil, fingerprint
}

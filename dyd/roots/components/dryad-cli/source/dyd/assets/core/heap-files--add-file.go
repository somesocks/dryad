package core

import (
	"os"
	"errors"
	"io/fs"
	"path/filepath"

	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

type heapAddFileRequest struct {
	HeapFiles *SafeHeapFilesReference
	SourcePath string
}

func heapAddFile(ctx *task.ExecutionContext, req heapAddFileRequest) (error, string) {
	var heapFilesPath string = req.HeapFiles.BasePath
	var sourcePath string = req.SourcePath
	var err error

	zlog.
		Trace().
		Str("heapFilesPath", heapFilesPath).
		Str("sourcePath", sourcePath).
		Msg("HeapAddFile")

	sourceHashAlgorithm, sourceHash, err := fileHash(sourcePath)
	if err != nil {
		return err, ""
	}

	fingerprint := sourceHashAlgorithm + "-" + sourceHash

	destPath := filepath.Join(heapFilesPath, fingerprint)

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

func (heapFiles *SafeHeapFilesReference) AddFile(
	ctx *task.ExecutionContext,
	sourcePath string,
) (error, string) {
	err, res := heapAddFile(
		ctx,
		heapAddFileRequest{
			HeapFiles: heapFiles,
			SourcePath: sourcePath,
		},
	)
	return err, res
}
package core

import (
	"dryad/internal/os"
	"dryad/task"
	"errors"
	"io/fs"
	stdos "os"
	"path/filepath"
	"time"

	zlog "github.com/rs/zerolog/log"
)

type heapAddFileRequest struct {
	HeapFiles  *SafeHeapFilesReference
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
	now := time.Now()

	// Fast path: if the CAS entry already exists, avoid unnecessary temp writes.
	if _, err := os.Stat(destPath); err == nil {
		err = stdos.Chtimes(destPath, now, now)
		if err != nil {
			zlog.Warn().
				Str("path", destPath).
				Err(err).
				Msg("failed to update heap file timestamps on existing entry")
		}
		return nil, fingerprint
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err, ""
	}

	srcFile, err := stdos.Open(sourcePath)
	if err != nil {
		return err, ""
	}
	defer srcFile.Close()

	tempFile, err := os.CreateTemp(
		heapFilesPath,
		".tmp-"+fingerprint+"-*",
	)
	if err != nil {
		return err, ""
	}
	tempPath := tempFile.Name()
	// Best effort cleanup. Crash/power-loss can still leave tmp files behind.
	defer os.Remove(tempPath)

	_, err = tempFile.ReadFrom(srcFile)
	if err != nil {
		return err, ""
	}

	// heap files should be set to R-X--X--X
	err = tempFile.Chmod(0o511)
	if err != nil {
		return err, ""
	}

	err = tempFile.Close()
	if err != nil {
		return err, ""
	}

	// Publish atomically without overwriting an existing CAS entry.
	err = os.Link(tempPath, destPath)
	if err != nil {
		// return a success if the file is already in the heap
		if errors.Is(err, fs.ErrExist) {
			err = stdos.Chtimes(destPath, now, now)
			if err != nil {
				zlog.Warn().
					Str("path", destPath).
					Err(err).
					Msg("failed to update heap file timestamps on existing entry")
			}
			return nil, fingerprint
		}
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
			HeapFiles:  heapFiles,
			SourcePath: sourcePath,
		},
	)
	return err, res
}

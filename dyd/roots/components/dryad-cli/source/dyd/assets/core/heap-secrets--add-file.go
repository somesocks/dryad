package core

import (
	"dryad/internal/os"
	"dryad/internal/time"
	"dryad/task"
	"errors"
	"io/fs"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type heapAddSecretFileRequest struct {
	HeapSecrets *SafeHeapSecretsReference
	SourcePath  string
}

func heapAddSecretFile(ctx *task.ExecutionContext, req heapAddSecretFileRequest) (error, string) {
	var heapSecretsPath string = req.HeapSecrets.BasePath
	var sourcePath string = req.SourcePath
	var err error

	zlog.
		Trace().
		Str("heapSecretsPath", heapSecretsPath).
		Str("sourcePath", sourcePath).
		Msg("HeapAddSecretFile")

	sourceHashAlgorithm, sourceHash, err := fileHash(sourcePath)
	if err != nil {
		return err, ""
	}

	fingerprint := sourceHashAlgorithm + "-" + sourceHash

	destPath := filepath.Join(heapSecretsPath, fingerprint)
	now := time.Now()

	// Fast path: if the CAS entry already exists, avoid unnecessary temp writes.
	if _, err := os.Stat(destPath); err == nil {
		err = os.Chtimes(destPath, now, now)
		if err != nil {
			zlog.Warn().
				Str("path", destPath).
				Err(err).
				Msg("failed to update heap secret timestamps on existing entry")
		}
		return nil, fingerprint
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err, ""
	}

	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return err, ""
	}
	defer srcFile.Close()

	tempFile, err := os.CreateTemp(
		heapSecretsPath,
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
			err = os.Chtimes(destPath, now, now)
			if err != nil {
				zlog.Warn().
					Str("path", destPath).
					Err(err).
					Msg("failed to update heap secret timestamps on existing entry")
			}
			return nil, fingerprint
		}
		return err, ""
	}

	return nil, fingerprint
}

func (heapSecrets *SafeHeapSecretsReference) AddFile(
	ctx *task.ExecutionContext,
	sourcePath string,
) (error, string) {
	err, res := heapAddSecretFile(
		ctx,
		heapAddSecretFileRequest{
			HeapSecrets: heapSecrets,
			SourcePath:  sourcePath,
		},
	)
	return err, res
}

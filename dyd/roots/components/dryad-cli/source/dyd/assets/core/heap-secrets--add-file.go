package core

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"dryad/task"

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
		}
		return err, ""
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

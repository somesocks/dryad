package core

import (
	"os"
	"errors"
	"io/fs"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

func HeapAddFile(heapPath string, filePath string) (string, error) {
	zlog.
		Trace().
		Str("heapPath", heapPath).
		Str("filePath", filePath).
		Msg("HeapAddFile")

	// fmt.Println("[trace] HeapAddFile", heapPath, filePath)
	heapPath, err := HeapPath(heapPath)
	if err != nil {
		return "", err
	}

	fileHashAlgorithm, fileHash, err := fileHash(filePath)
	if err != nil {
		return "", err
	}

	fingerprint := fileHashAlgorithm + "-" + fileHash

	destPath := filepath.Join(heapPath, "files", fingerprint)

	srcFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	var destFile *os.File
	destFile, err = os.OpenFile(destPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		// return a success if the file is already in the heap
		if errors.Is(err, fs.ErrExist) {
			return fingerprint, nil
		} else {
			return "", err
		}
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(srcFile)
	if err != nil {
		return "", err
	}

	// heap files should be set to R-X--X--X
	err = destFile.Chmod(0o511)
	if err != nil {
		return "", err
	}

	return fingerprint, nil
}

package core

import (
	"os"
	"path/filepath"
)

func HeapAddFile(heapPath string, filePath string) (string, error) {
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

	fileExists, err := fileExists(destPath)
	if err != nil {
		return "", err
	}

	if !fileExists {
		srcFile, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer srcFile.Close()

		var destFile *os.File
		destFile, err = os.Create(destPath)
		if err != nil {
			return "", err
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

		err = destFile.Sync()
		if err != nil {
			return "", err
		}
	}

	return fingerprint, nil
}

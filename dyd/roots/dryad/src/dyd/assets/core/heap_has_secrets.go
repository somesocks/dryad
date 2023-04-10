package core

import (
	"os"
	"path/filepath"
)

func HeapHasSecrets(path string, fingerprint string) (string, error) {
	// fmt.Println("[trace] HeapHasSecrets ", path, fingerprint)

	var heapPath, heapErr = HeapPath(path)
	if heapErr != nil {
		return "", heapErr
	}

	var stemPath = filepath.Join(heapPath, "secrets", fingerprint)
	var fileInfo, fileInfoErr = os.Stat(stemPath)

	if fileInfoErr == nil && fileInfo.IsDir() {
		return stemPath, nil
	} else {
		return "", nil
	}

}

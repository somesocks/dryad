package core

import (
	"os"
	"path/filepath"
)

func HeapHasSecrets(garden *SafeGardenReference, fingerprint string) (string, error) {
	// fmt.Println("[trace] HeapHasSecrets ", path, fingerprint)

	var heapPath, heapErr = HeapPath(garden)
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

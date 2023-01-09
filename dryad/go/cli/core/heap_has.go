package core

import (
	"os"
	"path/filepath"
)

func HeapHas(path string, fingerprint string) (bool, error) {
	var heapPath, heapErr = HeapPath(path)
	if heapErr != nil {
		return false, heapErr
	}

	var stemPath = filepath.Join(heapPath, fingerprint)
	var fileInfo, fileInfoErr = os.Stat(stemPath)

	if fileInfoErr == nil && fileInfo.IsDir() {
		return true, nil
	} else {
		return false, nil
	}

}

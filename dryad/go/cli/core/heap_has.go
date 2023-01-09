package core

import (
	"os"
	"path/filepath"
)

func HeapHas(path string, fingerprint string) (bool, error) {
	var heapPath, heapErr = HeapFind(path)
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

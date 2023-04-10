package core

import (
	"fmt"
	"os"
	"path/filepath"
)

func HeapHasStem(path string, fingerprint string) (string, error) {
	// fmt.Println("[trace] HeapHasStem ", path, fingerprint)

	var heapPath, heapErr = HeapPath(path)
	// fmt.Println("HeapHasStem heapPath ", heapPath)
	if heapErr != nil {
		return "", heapErr
	}

	var stemPath = filepath.Join(heapPath, "stems", fingerprint)
	var fileInfo, fileInfoErr = os.Stat(stemPath)

	if fileInfoErr == nil && fileInfo.IsDir() {
		return stemPath, nil
	} else {
		return "", nil
	}

}
